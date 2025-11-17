# Student Housing System — API Specification

Version: **1.0**
Format: **REST + JSON**

---

# Authentication & Authorization

### **POST /auth/login**

Authenticate student, housing staff, or admin.

**Body**

```json
{
  "email": "student@nu.edu.kz",
  "password": "password123"
}
```

**Response**

```json
{
  "token": "jwt_token",
  "role": "student"
}
```

---

### **POST /auth/register**

*(Admins only if staff creation is manual)*

**Body**

```json
{
  "nu_id": "202011111",
  "email": "example@nu.edu.kz",
  "role": "student | housing | admin"
}
```

---

### **POST /auth/change-password**

Authenticated users update their password.

**Body**

```json
{
  "old_password": "oldpass",
  "new_password": "newpass123"
}
```

---

# Student Endpoints

## 1. Application Management

### **POST /student/application**

Submit a new housing application.

**Body**

```json
{
  "phone": "87021234567",
  "year": 2,
  "major": "Computer Science",
  "gender": "male",
  "room_preference": "double",
  "additional_info": "No allergies"
}
```

---

### **GET /student/application**

Fetch student’s current application.

**Response**

```json
{
  "application_id": 12,
  "status": "pending",
  "submitted_at": "2025-02-15T12:31:00Z"
}
```

---

### **PATCH /student/application**

Edit application before submission period closes.

**Body**

```json
{
  "phone": "87023334444",
  "room_preference": "single"
}
```

---

### **DELETE /student/application**

Cancel the application.

---

## 2. Document Upload

### **POST /student/documents**

Upload a document.

**Form-Data**

* `file`: (PDF/JPG/PNG)
* `type`: "id_card" | "enrollment_certificate" | "photo" | etc.

**Response**

```json
{
  "document_id": 55,
  "url": "https://minio.example/doc1.pdf"
}
```

---

### **GET /student/documents**

List all submitted documents.

---

### **PATCH /student/documents/{id}**

Resubmit/update a document.

---

## 3. Notifications

### **GET /student/notifications**

List all notifications.

**Response**

```json
[ 
  {
    "id": 1,
    "message": "Your application was approved",
    "timestamp": "2025-02-20T09:01:00Z",
    "read": false
  }
]
```

---

### **GET /student/status**

Retrieve the latest status of application.

---

# Housing Staff Endpoints

## 1. Application Review

### **GET /housing/applications**

Query + filter applications.

**Query Params**

```
status=pending
year=2
gender=male
major=CS
search=Bakdaulet
```

---

### **GET /housing/applications/{id}**

Get full application details + documents.

---

### **PATCH /housing/applications/{id}/approve**

Approve application.

---

### **PATCH /housing/applications/{id}/reject**

Reject application.

**Body**

```json
{
  "reason": "Document was blurry"
}
```

---

## 2. Document Handling

### **GET /housing/documents/{id}**

View a specific document.

---

# Admin Endpoints

## 1. User Management (CRUD)

### **GET /admin/users**

List all users.

**Query params**

```
role=student|housing|admin
search=NameOrEmail
```

---

### **POST /admin/users**

Create new user.

---

### **PATCH /admin/users/{id}**

Modify user details.

---

### **DELETE /admin/users/{id}**

Deactivate user.

---

## 2. System Monitoring

### **GET /admin/system/health**

Returns health of all services.

**Response**

```json
{
  "backend": "OK",
  "database": "OK",
  "storage": "OK",
  "notifications": "OK"
}
```

---

### **GET /admin/system/stats**

Dashboard metrics.

**Response**

```json
{
  "total_students": 1200,
  "pending_applications": 310,
  "approved_applications": 820,
  "rejected_applications": 70
}
```

---

## 3. System-Wide Settings

### **GET /admin/settings**

Get current system settings.

---

### **PATCH /admin/settings**

Update settings.

**Body**

```json
{
  "application_open": "2025-08-01",
  "application_close": "2025-09-01",
  "required_documents": ["id_card", "enrollment_certificate"]
}
```

---

# Internal Services

## ## Notifications Service

### **POST /internal/notifications/send**

Used by backend services to send notifications.

**Body**

```json
{
  "user_id": 123,
  "type": "email",
  "message": "Your application was updated"
}
```

---

# File Storage (MinIO)

### **GET /internal/files/{path}**

Securely retrieve stored files.

---

# Response Codes

* **200 OK**
* **201 Created**
* **204 No Content**
* **400 Bad Request**
* **401 Unauthorized**
* **403 Forbidden**
* **404 Not Found**
* **500 Internal Server Error**

---

# Health & Debug

### **GET /health**

Check if backend is alive.
