package database

import "testing"

func TestBuildRoomListStartsFromSecondFloorAndAscendingBlocks(t *testing.T) {
	rooms := buildRoomList([]int{22}, 2)
	if len(rooms) == 0 {
		t.Fatal("expected rooms")
	}
	if rooms[0].Block != 22 || rooms[0].RoomNumber != 201 || rooms[0].Capacity != 2 {
		t.Fatalf("first room = %+v, want block 22 room 201 capacity 2", rooms[0])
	}
	last := rooms[len(rooms)-1]
	if last.RoomNumber != 1228 {
		t.Fatalf("last room number = %d, want 1228", last.RoomNumber)
	}
}

func TestParsePreferredRoommateIDs(t *testing.T) {
	got := parsePreferredRoommateIDs("preferred_roommate: 12, 34; 56")
	want := []int{12, 34, 56}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}

func TestAllocateApplicantsKeepsMutualRoommatesTogether(t *testing.T) {
	applicants := []allocationApplicant{
		{ApplicationID: 1, StudentID: 101, Gender: "female", Major: "CS", PreferredRoommateIDs: []int{102}},
		{ApplicationID: 2, StudentID: 102, Gender: "female", Major: "Math", PreferredRoommateIDs: []int{101}},
		{ApplicationID: 3, StudentID: 103, Gender: "female", Major: "Physics"},
	}

	assignments, err := allocateApplicants(applicants, map[string]roomOccupancy{})
	if err != nil {
		t.Fatalf("allocateApplicants returned error: %v", err)
	}
	if len(assignments) != 3 {
		t.Fatalf("assignments len = %d, want 3", len(assignments))
	}

	roomByApp := map[int]string{}
	for _, assignment := range assignments {
		roomByApp[assignment.ApplicationID] = roomKey(assignment.Block, assignment.RoomNumber)
	}
	if roomByApp[1] != roomByApp[2] {
		t.Fatalf("mutual roommates allocated to different rooms: %q vs %q", roomByApp[1], roomByApp[2])
	}
}

func TestAllocateApplicantsUsesNUFYPBlocksForNUFYPStudents(t *testing.T) {
	assignments, err := allocateApplicants([]allocationApplicant{
		{ApplicationID: 1, StudentID: 101, Gender: "male", Major: "NUFYP Engineering"},
	}, map[string]roomOccupancy{})
	if err != nil {
		t.Fatalf("allocateApplicants returned error: %v", err)
	}
	if len(assignments) != 1 {
		t.Fatalf("assignments len = %d, want 1", len(assignments))
	}
	block := assignments[0].Block
	if block != 11 && block != 19 && block != 20 {
		t.Fatalf("NUFYP student allocated to block %d, want 11/19/20", block)
	}
}

func TestAllocateApplicantsUsesFreeBedWhenExistingBedsHaveGap(t *testing.T) {
	assignments, err := allocateApplicants([]allocationApplicant{
		{ApplicationID: 1, StudentID: 101, Gender: "female", Major: "CS"},
	}, map[string]roomOccupancy{
		roomKey(22, 201): {
			Gender: "female",
			Count:  1,
			Beds:   map[int]bool{2: true},
		},
	})
	if err != nil {
		t.Fatalf("allocateApplicants returned error: %v", err)
	}
	if len(assignments) != 1 {
		t.Fatalf("assignments len = %d, want 1", len(assignments))
	}
	if assignments[0].Block != 22 || assignments[0].RoomNumber != 201 || assignments[0].BedNumber != 1 {
		t.Fatalf("assignment = %+v, want block 22 room 201 bed 1", assignments[0])
	}
}
