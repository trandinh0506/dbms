package repository

import (
	"context"
	"database/sql"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type Repository struct {
	DB *sql.DB
}

func (r *Repository) RegisterAccount(user, pass, name, role string) (int, error) {
	var id int
	err := r.DB.QueryRow("EXEC RegisterAccount @p1, @p2, @p3, @p4", user, pass, name, role).Scan(&id)
	return id, err
}

func (r *Repository) UpdateMark(ctx context.Context, tx *sql.Tx, sID int, cID int, mark float64) error {
    query := "UPDATE ENROLLMENT SET mark = @p1 WHERE student_id = @p2 AND course_id = @p3"
    _, err := tx.ExecContext(ctx, query, mark, sID, cID)
    return err
}

func (r *Repository) GetMark(ctx context.Context, tx *sql.Tx, sID int, cID int) (float64, error) {
    var mark float64
    query := "SELECT mark FROM ENROLLMENT WHERE student_id = @p1 AND course_id = @p2"
    err := tx.QueryRowContext(ctx, query, sID, cID).Scan(&mark)
    return mark, err
}

func (r *Repository) CreateUser(user, pass, name, role string) (int, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	var id int
	err = r.DB.QueryRow("EXEC RegisterAccount @p1, @p2, @p3, @p4",
		user, string(hashedBytes), name, role).Scan(&id)

	return id, err
}

func (r *Repository) GetAllCourses() ([]gin.H, error) {
	query := `SELECT CourseName, Lecturers, StudentCount FROM View_CourseLecturers`

	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []gin.H
	for rows.Next() {
		var name, lecturers string
		var count int
		rows.Scan(&name, &lecturers, &count)
		courses = append(courses, gin.H{
			"course_name": name,
			"lecturers":   lecturers,
			"students":    count,
		})
	}
	return courses, nil
}
