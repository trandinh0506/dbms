package repository

import (
	"context"
	"dbms/internal/models"
)

func (r *Repository) GetStudentGrades(studentID int) ([]models.StudentGrade, error) {
	rows, err := r.DB.Query("SELECT CourseName, mark, review, updated_at FROM View_StudentGrades WHERE StudentID = @p1", studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var grades []models.StudentGrade
	for rows.Next() {
		var g models.StudentGrade
		rows.Scan(&g.CourseName, &g.Mark, &g.Review, &g.UpdatedAt)
		grades = append(grades, g)
	}
	return grades, nil
}

func (r *Repository) GetGPA(studentID int) (float64, error) {
	var gpa float64
	err := r.DB.QueryRow("SELECT dbo.fn_GetStudentGPA(@p1)", studentID).Scan(&gpa)
	return gpa, err
}

func (r *Repository) EnrollCourse(studentID, courseID int) error {
	_, err := r.DB.Exec("EXEC EnrollStudent @p1, @p2", studentID, courseID)

	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetGradesUnsafe(ctx context.Context, studentID int) ([]struct {
	ID   int
	Name string
	Mark float64
}, error) {
	query := `
        SELECT c.id, c.name, e.mark
        FROM ENROLLMENT e WITH (NOLOCK)
        JOIN COURSE c WITH (NOLOCK) ON e.course_id = c.id
        WHERE e.student_id = @p1`

	rows, err := r.DB.QueryContext(ctx, query, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []struct {
		ID   int
		Name string
		Mark float64
	}
	for rows.Next() {
		var c struct {
			ID   int
			Name string
			Mark float64
		}
		if err := rows.Scan(&c.ID, &c.Name, &c.Mark); err != nil {
			return nil, err
		}
		courses = append(courses, c)
	}
	return courses, nil
}

func (r *Repository) GetGradesSafe(ctx context.Context, sID int) ([]struct {
	Name string
	Mark float64
}, error) {
	rows, err := r.DB.QueryContext(ctx, "EXEC GetStudentGradesSafe @p1", sID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []struct {
		Name string
		Mark float64
	}
	for rows.Next() {
		var res struct {
			Name string
			Mark float64
		}
		if err := rows.Scan(&res.Name, &res.Mark); err != nil {
			return nil, err
		}
		results = append(results, res)
	}
	return results, nil
}
