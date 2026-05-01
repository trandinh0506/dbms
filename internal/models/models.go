package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Account struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	Role         string `json:"role"`
}

type Student struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Balance float64 `json:"balance"`
}

type Instructor struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Enrollment struct {
	StudentID int       `json:"student_id"`
	CourseID  int       `json:"course_id"`
	Mark      float64   `json:"mark"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserClaims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type StudentGrade struct {
	CourseName string  `json:"course_name"`
	Mark       float64 `json:"mark"`
	Review     string  `json:"review"`
	UpdatedAt  string  `json:"updated_at"`
}
