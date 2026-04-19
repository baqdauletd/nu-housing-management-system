package database

import (
	"database/sql"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"

	"nu-housing-management-system/backend/internal/models"
)

type allocationApplicant struct {
	ApplicationID        int
	StudentID            int
	Gender               string
	Major                string
	PreferredRoommateIDs []int
}

type roomSlot struct {
	Block      int
	RoomNumber int
	Capacity   int
}

type roomOccupancy struct {
	Gender string
	Count  int
}

type pendingAllocation struct {
	ApplicationID int
	StudentID     int
	Block         int
	RoomNumber    int
	BedNumber     int
}

type ApplicationNotificationRecipient struct {
	ApplicationID  int
	StudentID      int
	Email          string
	StudentNumber  string
	DisplayName    string
	Status         string
	RejectedReason string
	Block          *int
	RoomNumber     *int
	BedNumber      *int
}

func RunRoomAllocation(db *sql.DB) error {
	if err := ensureRoomAllocationSchema(db); err != nil {
		return err
	}

	applicants, err := listApprovedApplicantsForAllocation(db)
	if err != nil {
		return err
	}
	if len(applicants) == 0 {
		return nil
	}

	occupancy, err := loadRoomOccupancy(db)
	if err != nil {
		return err
	}

	assignments, err := allocateApplicants(applicants, occupancy)
	if err != nil {
		return err
	}
	if len(assignments) == 0 {
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, assignment := range assignments {
		if _, err := tx.Exec(`
			INSERT INTO room_allocations (application_id, student_id, block, room_number, bed_number)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (application_id) DO NOTHING
		`, assignment.ApplicationID, assignment.StudentID, assignment.Block, assignment.RoomNumber, assignment.BedNumber); err != nil {
			return err
		}

	}

	return tx.Commit()
}

func ensureRoomAllocationSchema(db *sql.DB) error {
	statements := []string{
		`
		CREATE TABLE IF NOT EXISTS room_allocations (
			id SERIAL PRIMARY KEY,
			application_id INT UNIQUE REFERENCES applications(id) ON DELETE CASCADE,
			student_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			block INT NOT NULL,
			room_number INT NOT NULL,
			bed_number INT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			UNIQUE(block, room_number, bed_number)
		)
		`,
		`CREATE INDEX IF NOT EXISTS idx_room_allocations_student_id ON room_allocations(student_id)`,
		`CREATE INDEX IF NOT EXISTS idx_room_allocations_room ON room_allocations(block, room_number)`,
		`ALTER TABLE room_allocations ADD COLUMN IF NOT EXISTS ends_at TIMESTAMP NULL`,
	}

	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			return err
		}
	}
	return nil
}

type DormInventoryFilters struct {
	Block  *int
	Room   *int
	Gender string
	Major  string
	Search string
}

func ListDormInventory(db *sql.DB, filters DormInventoryFilters) ([]models.DormInventoryRow, error) {
	query := `
		SELECT
			ra.application_id,
			ra.student_id,
			u.nu_id,
			COALESCE(NULLIF(a.fio, ''), NULLIF(a.name_surname, ''), u.email),
			u.email,
			COALESCE(a.gender, ''),
			COALESCE(a.major, ''),
			ra.block,
			ra.room_number,
			ra.bed_number,
			ra.created_at,
			ra.ends_at
		FROM room_allocations ra
		JOIN applications a ON a.id = ra.application_id
		JOIN users u ON u.id = ra.student_id
		WHERE 1=1
	`

	args := make([]any, 0, 5)
	argIndex := 1
	if filters.Block != nil {
		query += fmt.Sprintf(" AND ra.block = $%d", argIndex)
		args = append(args, *filters.Block)
		argIndex++
	}
	if filters.Room != nil {
		query += fmt.Sprintf(" AND ra.room_number = $%d", argIndex)
		args = append(args, *filters.Room)
		argIndex++
	}
	if gender := strings.TrimSpace(filters.Gender); gender != "" {
		query += fmt.Sprintf(" AND LOWER(COALESCE(a.gender, '')) = LOWER($%d)", argIndex)
		args = append(args, gender)
		argIndex++
	}
	if major := strings.TrimSpace(filters.Major); major != "" {
		query += fmt.Sprintf(" AND COALESCE(a.major, '') ILIKE $%d", argIndex)
		args = append(args, "%"+major+"%")
		argIndex++
	}
	if search := strings.TrimSpace(filters.Search); search != "" {
		query += fmt.Sprintf(` AND (
			u.nu_id ILIKE $%d OR
			u.email ILIKE $%d OR
			COALESCE(a.fio, '') ILIKE $%d OR
			COALESCE(a.name_surname, '') ILIKE $%d OR
			CAST(ra.block AS TEXT) ILIKE $%d OR
			CAST(ra.room_number AS TEXT) ILIKE $%d
		)`, argIndex, argIndex, argIndex, argIndex, argIndex, argIndex)
		args = append(args, "%"+search+"%")
		argIndex++
	}
	query += " ORDER BY ra.block ASC, ra.room_number ASC, ra.bed_number ASC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]models.DormInventoryRow, 0)
	byRoom := make(map[string][]models.DormInventoryRow)
	for rows.Next() {
		var row models.DormInventoryRow
		if err := rows.Scan(
			&row.ApplicationID,
			&row.StudentID,
			&row.NuID,
			&row.FIO,
			&row.Email,
			&row.Gender,
			&row.Major,
			&row.Block,
			&row.RoomNumber,
			&row.BedNumber,
			&row.FromWhen,
			&row.TillWhen,
		); err != nil {
			return nil, err
		}
		row.Capacity = roomCapacityForBlock(row.Block)
		row.Roommates = []models.DormInventoryRoommate{}
		result = append(result, row)
		byRoom[roomKey(row.Block, row.RoomNumber)] = append(byRoom[roomKey(row.Block, row.RoomNumber)], row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for idx := range result {
		roommates := byRoom[roomKey(result[idx].Block, result[idx].RoomNumber)]
		result[idx].OccupancyCount = len(roommates)
		result[idx].Roommates = make([]models.DormInventoryRoommate, 0, max(0, len(roommates)-1))
		for _, roommate := range roommates {
			if roommate.ApplicationID == result[idx].ApplicationID {
				continue
			}
			result[idx].Roommates = append(result[idx].Roommates, models.DormInventoryRoommate{
				ApplicationID: roommate.ApplicationID,
				StudentID:     roommate.StudentID,
				NuID:          roommate.NuID,
				FIO:           roommate.FIO,
				Email:         roommate.Email,
				BedNumber:     roommate.BedNumber,
			})
		}
	}

	return result, nil
}

func GetRoomAllocationByApplicationID(db *sql.DB, applicationID int) (models.RoomAllocation, error) {
	var allocation models.RoomAllocation
	err := db.QueryRow(`
		SELECT id, application_id, student_id, block, room_number, bed_number, created_at
		FROM room_allocations
		WHERE application_id = $1
	`, applicationID).Scan(
		&allocation.ID,
		&allocation.ApplicationID,
		&allocation.StudentID,
		&allocation.Block,
		&allocation.RoomNumber,
		&allocation.BedNumber,
		&allocation.CreatedAt,
	)
	return allocation, err
}

func ListRoomAllocationsByApplicationIDs(db *sql.DB, applicationIDs []int) (map[int]models.RoomAllocation, error) {
	result := make(map[int]models.RoomAllocation, len(applicationIDs))
	if len(applicationIDs) == 0 {
		return result, nil
	}

	rows, err := db.Query(`
		SELECT id, application_id, student_id, block, room_number, bed_number, created_at
		FROM room_allocations
		WHERE application_id = ANY($1)
	`, pq.Array(applicationIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var allocation models.RoomAllocation
		if err := rows.Scan(
			&allocation.ID,
			&allocation.ApplicationID,
			&allocation.StudentID,
			&allocation.Block,
			&allocation.RoomNumber,
			&allocation.BedNumber,
			&allocation.CreatedAt,
		); err != nil {
			return nil, err
		}
		result[allocation.ApplicationID] = allocation
	}
	return result, rows.Err()
}

func ListApplicationNotificationRecipients(db *sql.DB) ([]ApplicationNotificationRecipient, error) {
	rows, err := db.Query(`
		SELECT
			a.id,
			a.student_id,
			u.email,
			COALESCE(NULLIF(a.student_number, ''), u.nu_id),
			COALESCE(NULLIF(a.name_surname, ''), NULLIF(a.fio, ''), u.email),
			COALESCE(a.status, ''),
			COALESCE(a.rejected_reason, ''),
			ra.block,
			ra.room_number,
			ra.bed_number
		FROM applications a
		JOIN users u ON u.id = a.student_id
		LEFT JOIN room_allocations ra ON ra.application_id = a.id
		WHERE COALESCE(a.status, '') <> 'canceled'
		ORDER BY a.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipients []ApplicationNotificationRecipient
	for rows.Next() {
		var recipient ApplicationNotificationRecipient
		var block, roomNumber, bedNumber sql.NullInt64
		if err := rows.Scan(
			&recipient.ApplicationID,
			&recipient.StudentID,
			&recipient.Email,
			&recipient.StudentNumber,
			&recipient.DisplayName,
			&recipient.Status,
			&recipient.RejectedReason,
			&block,
			&roomNumber,
			&bedNumber,
		); err != nil {
			return nil, err
		}
		if block.Valid {
			value := int(block.Int64)
			recipient.Block = &value
		}
		if roomNumber.Valid {
			value := int(roomNumber.Int64)
			recipient.RoomNumber = &value
		}
		if bedNumber.Valid {
			value := int(bedNumber.Int64)
			recipient.BedNumber = &value
		}
		recipients = append(recipients, recipient)
	}
	return recipients, rows.Err()
}

func InsertApplicationClosureNotifications(db *sql.DB, recipients []ApplicationNotificationRecipient) error {
	if len(recipients) == 0 {
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, recipient := range recipients {
		message := ApplicationClosureNotificationMessage(recipient)
		if _, err := tx.Exec(
			`INSERT INTO notifications (user_id, message, read, created_at) VALUES ($1, $2, FALSE, NOW())`,
			recipient.StudentID,
			message,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func ApplicationClosureNotificationMessage(recipient ApplicationNotificationRecipient) string {
	switch strings.ToLower(strings.TrimSpace(recipient.Status)) {
	case "approved":
		if recipient.Block != nil && recipient.RoomNumber != nil && recipient.BedNumber != nil {
			return fmt.Sprintf("Your housing application was approved. Your room is %d.%d, bed %d.", *recipient.Block, *recipient.RoomNumber, *recipient.BedNumber)
		}
		return "Your housing application was approved. Your room allocation is being finalized."
	case "rejected":
		reason := strings.TrimSpace(recipient.RejectedReason)
		if reason == "" {
			return "Your housing application was rejected."
		}
		return fmt.Sprintf("Your housing application was rejected. Reason: %s", reason)
	default:
		return fmt.Sprintf("Your housing application status is %s.", strings.TrimSpace(recipient.Status))
	}
}

func listApprovedApplicantsForAllocation(db *sql.DB) ([]allocationApplicant, error) {
	rows, err := db.Query(`
		SELECT a.id, a.student_id, COALESCE(a.gender, ''), COALESCE(a.major, ''), COALESCE(a.additional_info, '')
		FROM applications a
		LEFT JOIN room_allocations ra ON ra.application_id = a.id
		WHERE a.status = 'approved'
		  AND ra.application_id IS NULL
		ORDER BY a.submitted_at ASC, a.id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var applicants []allocationApplicant
	for rows.Next() {
		var app allocationApplicant
		var additionalInfo string
		if err := rows.Scan(&app.ApplicationID, &app.StudentID, &app.Gender, &app.Major, &additionalInfo); err != nil {
			return nil, err
		}
		app.Gender = strings.ToLower(strings.TrimSpace(app.Gender))
		app.PreferredRoommateIDs = parsePreferredRoommateIDs(additionalInfo)
		applicants = append(applicants, app)
	}
	return applicants, rows.Err()
}

func loadRoomOccupancy(db *sql.DB) (map[string]roomOccupancy, error) {
	rows, err := db.Query(`
		SELECT ra.block, ra.room_number, COUNT(*)::int, COALESCE(LOWER(MAX(a.gender)), '')
		FROM room_allocations ra
		JOIN applications a ON a.id = ra.application_id
		GROUP BY ra.block, ra.room_number
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	occupancy := make(map[string]roomOccupancy)
	for rows.Next() {
		var block, roomNumber, count int
		var gender string
		if err := rows.Scan(&block, &roomNumber, &count, &gender); err != nil {
			return nil, err
		}
		occupancy[roomKey(block, roomNumber)] = roomOccupancy{Gender: gender, Count: count}
	}
	return occupancy, rows.Err()
}

func allocateApplicants(applicants []allocationApplicant, occupancy map[string]roomOccupancy) ([]pendingAllocation, error) {
	standardRooms := buildRoomList([]int{22, 23, 24, 25, 26, 27}, 2)
	nufypRooms := buildRoomList([]int{11, 19, 20}, 4)

	byCategoryGender := map[string][]allocationApplicant{}
	for _, applicant := range applicants {
		category := "standard"
		if isNUFYPMajor(applicant.Major) {
			category = "nufyp"
		}
		key := category + ":" + normalizeAllocationGender(applicant.Gender)
		byCategoryGender[key] = append(byCategoryGender[key], applicant)
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	assignments := make([]pendingAllocation, 0, len(applicants))
	for key, groupApplicants := range byCategoryGender {
		roomList := standardRooms
		if strings.HasPrefix(key, "nufyp:") {
			roomList = nufypRooms
		}
		groupAssignments, err := allocateGroup(groupApplicants, roomList, occupancy, rng)
		if err != nil {
			return nil, err
		}
		assignments = append(assignments, groupAssignments...)
	}

	return assignments, nil
}

func allocateGroup(applicants []allocationApplicant, rooms []roomSlot, occupancy map[string]roomOccupancy, rng *rand.Rand) ([]pendingAllocation, error) {
	if len(applicants) == 0 {
		return nil, nil
	}

	remaining := make([]allocationApplicant, len(applicants))
	copy(remaining, applicants)
	rng.Shuffle(len(remaining), func(i, j int) { remaining[i], remaining[j] = remaining[j], remaining[i] })

	byStudentID := make(map[int]allocationApplicant, len(remaining))
	for _, applicant := range remaining {
		byStudentID[applicant.StudentID] = applicant
	}

	used := make(map[int]bool, len(remaining))
	assignments := make([]pendingAllocation, 0, len(remaining))
	sort.SliceStable(remaining, func(i, j int) bool { return remaining[i].ApplicationID < remaining[j].ApplicationID })

	for _, applicant := range remaining {
		if used[applicant.ApplicationID] {
			continue
		}

		group := []allocationApplicant{applicant}
		if preferred := findMutualPreferredRoommate(applicant, byStudentID, used); preferred != nil {
			group = append(group, *preferred)
			used[preferred.ApplicationID] = true
		}
		used[applicant.ApplicationID] = true

		room, bedStart, err := findFirstAvailableRoom(group, rooms, occupancy)
		if err != nil {
			return nil, err
		}
		key := roomKey(room.Block, room.RoomNumber)
		current := occupancy[key]
		if current.Gender == "" && len(group) > 0 {
			current.Gender = normalizeAllocationGender(group[0].Gender)
		}

		for idx, member := range group {
			assignments = append(assignments, pendingAllocation{
				ApplicationID: member.ApplicationID,
				StudentID:     member.StudentID,
				Block:         room.Block,
				RoomNumber:    room.RoomNumber,
				BedNumber:     bedStart + idx,
			})
		}
		current.Count += len(group)
		occupancy[key] = current
	}

	return assignments, nil
}

func findMutualPreferredRoommate(applicant allocationApplicant, byStudentID map[int]allocationApplicant, used map[int]bool) *allocationApplicant {
	for _, preferredID := range applicant.PreferredRoommateIDs {
		other, ok := byStudentID[preferredID]
		if !ok || used[other.ApplicationID] {
			continue
		}
		if normalizeAllocationGender(other.Gender) != normalizeAllocationGender(applicant.Gender) {
			continue
		}
		if isNUFYPMajor(other.Major) != isNUFYPMajor(applicant.Major) {
			continue
		}
		if containsInt(other.PreferredRoommateIDs, applicant.StudentID) {
			return &other
		}
	}
	return nil
}

func findFirstAvailableRoom(group []allocationApplicant, rooms []roomSlot, occupancy map[string]roomOccupancy) (roomSlot, int, error) {
	groupGender := ""
	if len(group) > 0 {
		groupGender = normalizeAllocationGender(group[0].Gender)
	}

	for _, room := range rooms {
		key := roomKey(room.Block, room.RoomNumber)
		current := occupancy[key]
		if current.Gender != "" && current.Gender != groupGender {
			continue
		}
		if current.Count+len(group) > room.Capacity {
			continue
		}
		return room, current.Count + 1, nil
	}

	if len(group) == 0 {
		return roomSlot{}, 0, fmt.Errorf("no room available")
	}
	return roomSlot{}, 0, fmt.Errorf("no room available for %s applicants", groupGender)
}

func buildRoomList(blocks []int, capacity int) []roomSlot {
	rooms := make([]roomSlot, 0, len(blocks)*11*28)
	for _, block := range blocks {
		for floor := 2; floor <= 12; floor++ {
			for room := 1; room <= 28; room++ {
				roomNumber := floor*100 + room
				rooms = append(rooms, roomSlot{
					Block:      block,
					RoomNumber: roomNumber,
					Capacity:   capacity,
				})
			}
		}
	}
	return rooms
}

func parsePreferredRoommateIDs(additionalInfo string) []int {
	for _, line := range strings.Split(additionalInfo, "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		if strings.ToLower(strings.TrimSpace(key)) != "preferred_roommate" {
			continue
		}
		fields := strings.FieldsFunc(value, func(r rune) bool {
			return r == ',' || r == ';' || r == ' ' || r == '\t' || r == '\n'
		})
		result := make([]int, 0, len(fields))
		for _, field := range fields {
			field = strings.TrimSpace(field)
			if field == "" {
				continue
			}
			id, err := strconv.Atoi(field)
			if err == nil {
				result = append(result, id)
			}
		}
		return result
	}
	return nil
}

func normalizeAllocationGender(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "male", "m":
		return "male"
	case "female", "f":
		return "female"
	default:
		return value
	}
}

func isNUFYPMajor(major string) bool {
	return strings.Contains(strings.ToUpper(strings.TrimSpace(major)), "NUFYP")
}

func containsInt(values []int, target int) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func roomKey(block, roomNumber int) string {
	return fmt.Sprintf("%d:%d", block, roomNumber)
}

func roomCapacityForBlock(block int) int {
	switch block {
	case 11, 19, 20:
		return 4
	default:
		return 2
	}
}
