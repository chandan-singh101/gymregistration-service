#!/bin/bash

API_URL="http://localhost:8000/register"

for i in $(seq 1 1000); do
  full_name="User$i Test"
  gender=$((RANDOM % 2))  # 0 -> Male, 1 -> Female
  if [ $gender -eq 0 ]; then
    gender="Male"
  else
    gender="Female"
  fi

  dob_year=$((1980 + RANDOM % 25))
  dob_month=$(printf "%02d" $((1 + RANDOM % 12)))
  dob_day=$(printf "%02d" $((1 + RANDOM % 28)))
  date_of_birth="$dob_year-$dob_month-$dob_day"

  phone_number="9$((RANDOM % 900000000 + 100000000))"
  email="user$i@example.com"
  address="House No $i, Random Street, City"
  height_cm=$((150 + RANDOM % 50))
  weight_kg=$((50 + RANDOM % 50))
  medical_conditions="None"
  fitness_goal="Fitness Goal $i"

  emergency_name="Emergency$i Person"
  emergency_relation="Friend"
  emergency_phone="9$((RANDOM % 900000000 + 100000000))"

  # Membership
  plan=("Monthly" "Quarterly" "Half-Yearly" "Yearly")
  membership_plan=${plan[$RANDOM % 4]}
  joining_date="2025-09-05"
  expiry_date="2026-09-05"
  fee_amount=$((1000 + RANDOM % 15000))
  fee_status=("Paid" "Pending")
  fee_status_val=${fee_status[$RANDOM % 2]}

  # JSON payload
  payload=$(cat <<EOF
{
  "full_name": "$full_name",
  "gender": "$gender",
  "date_of_birth": "$date_of_birth",
  "phone_number": "$phone_number",
  "email": "$email",
  "address": "$address",
  "height_cm": $height_cm,
  "weight_kg": $weight_kg,
  "medical_conditions": "$medical_conditions",
  "fitness_goal": "$fitness_goal",
  "emergency_contact": {
    "name": "$emergency_name",
    "relation": "$emergency_relation",
    "phone_number": "$emergency_phone"
  },
  "membership": {
    "plan": "$membership_plan",
    "joining_date": "$joining_date",
    "expiry_date": "$expiry_date",
    "fee_amount": $fee_amount,
    "fee_status": "$fee_status_val"
  }
}
EOF
)

  # Call API
  curl -s -o /dev/null -w "Inserted $i -> Status: %{http_code}\n" \
    -X POST "$API_URL" \
    -H "Content-Type: application/json" \
    -d "$payload"

done
