package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// Database connection details for Supabase
const (
	connectionString = "postgresql://postgres.ykqrfpbkxnnohffjdbdt:Chandan1%40singh@aws-1-ap-southeast-1.pooler.supabase.com:5432/postgres?sslmode=require"
)

// Member struct represents a gym member
type Member struct {
	ID                int64           `json:"id"`
	FullName          string          `json:"full_name"`
	Gender            string          `json:"gender"`
	DateOfBirth       string          `json:"date_of_birth"`
	PhoneNumber       string          `json:"phone_number"`
	Email             string          `json:"email"`
	Address           string          `json:"address"`
	HeightCm          float64         `json:"height_cm"`
	WeightKg          float64         `json:"weight_kg"`
	MedicalConditions string          `json:"medical_conditions"`
	FitnessGoal       string          `json:"fitness_goal"`
	EmergencyContact  EmergencyContact `json:"emergency_contact"`
	Membership        Membership      `json:"membership"`
}

// EmergencyContact struct represents emergency contact information
type EmergencyContact struct {
	Name        string `json:"name"`
	Relation    string `json:"relation"`
	PhoneNumber string `json:"phone_number"`
}

// Membership struct represents membership information
type Membership struct {
	Plan        string  `json:"plan"`
	JoiningDate string  `json:"joining_date"`
	ExpiryDate  string  `json:"expiry_date"`
	FeeAmount   float64 `json:"fee_amount"`
	FeeStatus   string  `json:"fee_status"`
}

var db *sql.DB

func main() {
	// Initialize database connection
	var err error
	db, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Verify database connection
	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("Successfully connected to database!")

	// Initialize router
	router := mux.NewRouter()

	// Define routes
	router.HandleFunc("/register", registerMember).Methods("POST")
	router.HandleFunc("/members/{id}", getMember).Methods("GET")
	router.HandleFunc("/members", getMembers).Methods("GET")
	router.HandleFunc("/members/{id}", updateMember).Methods("PUT")
	router.HandleFunc("/members/{id}", deleteMember).Methods("DELETE")
	router.HandleFunc("/members/{id}/field", updateMemberField).Methods("PATCH")

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	fmt.Printf("Server listening on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

// registerMember handles new member registration
func registerMember(w http.ResponseWriter, r *http.Request) {
	var member Member
	err := json.NewDecoder(r.Body).Decode(&member)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Insert into members table
	var memberID int64
	err = tx.QueryRow(`
		INSERT INTO members (full_name, gender, date_of_birth, phone_number, email, address, height_cm, weight_kg, medical_conditions, fitness_goal)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`, member.FullName, member.Gender, member.DateOfBirth, member.PhoneNumber, member.Email,
		member.Address, member.HeightCm, member.WeightKg, member.MedicalConditions, member.FitnessGoal).Scan(&memberID)

	if err != nil {
		tx.Rollback()
		http.Error(w, "Error creating member: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Insert emergency contact
	_, err = tx.Exec(`
		INSERT INTO emergency_contacts (member_id, name, relation, phone_number)
		VALUES ($1, $2, $3, $4)
	`, memberID, member.EmergencyContact.Name, member.EmergencyContact.Relation, member.EmergencyContact.PhoneNumber)

	if err != nil {
		tx.Rollback()
		http.Error(w, "Error adding emergency contact: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Determine plan ID based on plan name
	var planID int
	err = tx.QueryRow("SELECT id FROM membership_plans WHERE name = $1", member.Membership.Plan).Scan(&planID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Invalid membership plan: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Insert membership details
	_, err = tx.Exec(`
		INSERT INTO memberships (member_id, plan_id, joining_date, expiry_date, fee_amount, fee_status)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, memberID, planID, member.Membership.JoiningDate, member.Membership.ExpiryDate,
		member.Membership.FeeAmount, member.Membership.FeeStatus)

	if err != nil {
		tx.Rollback()
		http.Error(w, "Error creating membership: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		http.Error(w, "Error committing transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "Member registered successfully",
		"member_id": memberID,
	})
}

// getMember retrieves a specific member by ID
func getMember(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	var member Member
	err := db.QueryRow(`
		SELECT m.id, m.full_name, m.gender, m.date_of_birth, m.phone_number, 
				   m.email, m.address, m.height_cm, m.weight_kg, 
				   m.medical_conditions, m.fitness_goal,
				   ec.name, ec.relation, ec.phone_number,
				   mp.name, ms.joining_date, ms.expiry_date, ms.fee_amount, ms.fee_status
		FROM members m
		LEFT JOIN emergency_contacts ec ON m.id = ec.member_id
		LEFT JOIN memberships ms ON m.id = ms.member_id
		LEFT JOIN membership_plans mp ON ms.plan_id = mp.id
		WHERE m.id = $1
	`, id).Scan(
		&member.ID, &member.FullName, &member.Gender, &member.DateOfBirth, &member.PhoneNumber,
		&member.Email, &member.Address, &member.HeightCm, &member.WeightKg,
		&member.MedicalConditions, &member.FitnessGoal,
		&member.EmergencyContact.Name, &member.EmergencyContact.Relation, &member.EmergencyContact.PhoneNumber,
		&member.Membership.Plan, &member.Membership.JoiningDate, &member.Membership.ExpiryDate,
		&member.Membership.FeeAmount, &member.Membership.FeeStatus,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Member not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(member)
}

// getMembers retrieves all members
func getMembers(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`
		SELECT m.id, m.full_name, m.phone_number, m.email, mp.name, ms.fee_status
		FROM members m
		LEFT JOIN memberships ms ON m.id = ms.member_id
		LEFT JOIN membership_plans mp ON ms.plan_id = mp.id
		ORDER BY m.id
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var members []map[string]interface{}
	for rows.Next() {
		var id int64
		var fullName, phoneNumber, email, planName, feeStatus string

		err := rows.Scan(&id, &fullName, &phoneNumber, &email, &planName, &feeStatus)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		member := map[string]interface{}{
			"id":           id,
			"full_name":    fullName,
			"phone_number": phoneNumber,
			"email":        email,
			"plan":         planName,
			"fee_status":   feeStatus,
		}

		members = append(members, member)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

// updateMember updates a member's details
func updateMember(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	memberID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid member ID", http.StatusBadRequest)
		return
	}

	var member Member
	err = json.NewDecoder(r.Body).Decode(&member)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update members table
	_, err = tx.Exec(`
		UPDATE members 
		SET full_name = $1, gender = $2, date_of_birth = $3, phone_number = $4, 
			email = $5, address = $6, height_cm = $7, weight_kg = $8, 
			medical_conditions = $9, fitness_goal = $10
		WHERE id = $11
	`, member.FullName, member.Gender, member.DateOfBirth, member.PhoneNumber,
		member.Email, member.Address, member.HeightCm, member.WeightKg,
		member.MedicalConditions, member.FitnessGoal, memberID)

	if err != nil {
		tx.Rollback()
		http.Error(w, "Error updating member: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update emergency contact
	_, err = tx.Exec(`
		UPDATE emergency_contacts 
		SET name = $1, relation = $2, phone_number = $3
		WHERE member_id = $4
	`, member.EmergencyContact.Name, member.EmergencyContact.Relation, 
	   member.EmergencyContact.PhoneNumber, memberID)

	if err != nil {
		tx.Rollback()
		http.Error(w, "Error updating emergency contact: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Determine plan ID based on plan name
	var planID int
	err = tx.QueryRow("SELECT id FROM membership_plans WHERE name = $1", member.Membership.Plan).Scan(&planID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Invalid membership plan: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Update membership details
	_, err = tx.Exec(`
		UPDATE memberships 
		SET plan_id = $1, joining_date = $2, expiry_date = $3, fee_amount = $4, fee_status = $5
		WHERE member_id = $6
	`, planID, member.Membership.JoiningDate, member.Membership.ExpiryDate,
		member.Membership.FeeAmount, member.Membership.FeeStatus, memberID)

	if err != nil {
		tx.Rollback()
		http.Error(w, "Error updating membership: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		http.Error(w, "Error committing transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Member updated successfully",
	})
}

// deleteMember deletes a member
func deleteMember(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	memberID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid member ID", http.StatusBadRequest)
		return
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Delete from memberships first (due to foreign key constraints)
	_, err = tx.Exec("DELETE FROM memberships WHERE member_id = $1", memberID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Error deleting membership: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Delete from emergency_contacts
	_, err = tx.Exec("DELETE FROM emergency_contacts WHERE member_id = $1", memberID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Error deleting emergency contact: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Delete from members
	_, err = tx.Exec("DELETE FROM members WHERE id = $1", memberID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Error deleting member: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		http.Error(w, "Error committing transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Member deleted successfully",
	})
}

// updateMemberField updates a specific field of a member
func updateMemberField(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	memberID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid member ID", http.StatusBadRequest)
		return
	}

	// Parse the request body to get the field to update and its new value
	var updateData map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&updateData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if field and value are provided
	if len(updateData) != 1 {
		http.Error(w, "Only one field can be updated at a time", http.StatusBadRequest)
		return
	}

	var field string
	var value interface{}
	for k, v := range updateData {
		field = k
		value = v
		break
	}

	// List of allowed fields that can be updated
	allowedFields := map[string]bool{
		"full_name": true, "gender": true, "date_of_birth": true, 
		"phone_number": true, "email": true, "address": true, 
		"height_cm": true, "weight_kg": true, "medical_conditions": true, 
		"fitness_goal": true,
	}

	// Check if the field is allowed
	if !allowedFields[field] {
		http.Error(w, "Field not allowed to be updated", http.StatusBadRequest)
		return
	}

	// Update the field in the database
	_, err = db.Exec(fmt.Sprintf("UPDATE members SET %s = $1 WHERE id = $2", field), value, memberID)
	if err != nil {
		http.Error(w, "Error updating field: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Field updated successfully",
	})
}
