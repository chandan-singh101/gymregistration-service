# 🏋️ Gym Member REST API

This is a Go-based RESTful API for managing gym members. It connects to a PostgreSQL database (specifically, a Supabase instance) and provides a set of endpoints to handle member registration, retrieval, updates, and deletion.

---

## 🚀 Features
- **Member Management**: Create, retrieve, update, and delete gym member records.
- **Structured Data**: Member data is organized across multiple tables (`members`, `emergency_contacts`, and `memberships`) and managed within a single API.
- **Database Transactions**: All create, update, and delete operations are wrapped in database transactions to ensure data integrity.
- **Partial Updates**: A `PATCH` endpoint allows for updating a single field of a member's record.
- **Gorilla Mux Router**: Uses the `gorilla/mux` package for clean and organized routing.

---

## 📋 Prerequisites
- **Go**: You need Go installed on your system.
- **PostgreSQL Database**: This application requires a PostgreSQL database. The code is configured to work with a Supabase database instance, but you can change the connection string to any PostgreSQL database.
- **Go Modules**: The project uses the `gorilla/mux` and `lib/pq` Go packages.

---

## 🗄️ Database Setup

The application expects a PostgreSQL database with the following tables.  

### `membership_plans` table
```sql
CREATE TABLE membership_plans (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    price DECIMAL(10, 2)
);
```

### `members` table
```sql
CREATE TABLE members (
    id SERIAL PRIMARY KEY,
    full_name VARCHAR(255) NOT NULL,
    gender VARCHAR(10),
    date_of_birth DATE,
    phone_number VARCHAR(20) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    address TEXT,
    height_cm DECIMAL(5, 2),
    weight_kg DECIMAL(5, 2),
    medical_conditions TEXT,
    fitness_goal TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### `emergency_contacts` table
```sql
CREATE TABLE emergency_contacts (
    id SERIAL PRIMARY KEY,
    member_id INT UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    relation VARCHAR(50),
    phone_number VARCHAR(20),
    FOREIGN KEY (member_id) REFERENCES members (id) ON DELETE CASCADE
);
```

### `memberships` table
```sql
CREATE TABLE memberships (
    id SERIAL PRIMARY KEY,
    member_id INT UNIQUE NOT NULL,
    plan_id INT NOT NULL,
    joining_date DATE,
    expiry_date DATE,
    fee_amount DECIMAL(10, 2),
    fee_status VARCHAR(50),
    FOREIGN KEY (member_id) REFERENCES members (id) ON DELETE CASCADE,
    FOREIGN KEY (plan_id) REFERENCES membership_plans (id) ON DELETE RESTRICT
);
```

---

## ⚙️ Configure Database Connection

Open `main.go` and update the connection string with your PostgreSQL credentials:

```go
const (
    connectionString = "postgresql://your_user:your_password@your_host:your_port/your_db?sslmode=require"
)
```

If you are using Supabase, the format of the connection string will be similar to the one provided in the source code.

---

## 📦 Install Dependencies

```bash
go get github.com/gorilla/mux
go get github.com/lib/pq
```

---

## ▶️ Run the Application

```bash
go run main.go
```

The server will start on **port 8000** by default.  
You can change the port by setting the `PORT` environment variable.

---

## 📡 API Endpoints

### **POST /register**
Registers a new gym member.  
- Request Body: JSON object with the Member struct.  
- Response: `201 Created` on success.

---

### **GET /members/{id}**
Retrieves a single member by their ID.  
- URL Parameter: `id` (integer)  
- Response: `200 OK` with JSON object, or `404 Not Found`

---

### **GET /members**
Retrieves a list of all members.  
- Response: `200 OK` with JSON array

---

### **PUT /members/{id}**
Updates all details for an existing member.  
- URL Parameter: `id` (integer)  
- Request Body: JSON object with updated member data  
- Response: `200 OK`

---

### **DELETE /members/{id}**
Deletes a member and their associated records.  
- URL Parameter: `id` (integer)  
- Response: `200 OK`

---

### **PATCH /members/{id}/field**
Updates a single field of a member's record.  
- URL Parameter: `id` (integer)  
- Request Body: JSON object containing a single key-value pair.  

Example:
```json
{ "email": "new.email@example.com" }
```

- Response: `200 OK`

---
