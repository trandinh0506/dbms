package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
)

func (r *Repository) GetCoursesByInstructor(instructorID int) ([]gin.H, error) {
	query := `
        SELECT c.id, v.CourseName, v.Lecturers, v.StudentCount 
        FROM View_CourseLecturers v
        JOIN COURSE c ON v.CourseName = c.name
        JOIN COURSE_INSTRUCTOR ci ON c.id = ci.course_id
        WHERE ci.instructor_id = @p1`

	rows, err := r.DB.Query(query, instructorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []gin.H
	for rows.Next() {
		var id int
		var name, lecturers string
		var count int
		if err := rows.Scan(&id, &name, &lecturers, &count); err != nil {
			continue
		}
		courses = append(courses, gin.H{
			"course_id":   id,
			"course_name": name,
			"lecturers":   lecturers,
			"students":    count,
		})
	}
	return courses, nil
}

func (r *Repository) CreateCourse(name string, capacity int, instructorID int) (int, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return 0, err
	}

	var courseID int
	queryCourse := "INSERT INTO COURSE (name, capacity, created_by) OUTPUT INSERTED.id VALUES (@p1, @p2, @p3)"
	err = tx.QueryRow(queryCourse, name, capacity, instructorID).Scan(&courseID)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	queryAssign := "INSERT INTO COURSE_INSTRUCTOR (course_id, instructor_id) VALUES (@p1, @p2)"
	_, err = tx.Exec(queryAssign, courseID, instructorID)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	err = tx.Commit()
	return courseID, err
}

func (r *Repository) InviteInstructor(courseID int, instructorID int) error {
	var exists bool
	err := r.DB.QueryRow("SELECT CASE WHEN EXISTS(SELECT 1 FROM COURSE_INSTRUCTOR WHERE course_id = @p1 AND instructor_id = @p2) THEN 1 ELSE 0 END",
		courseID, instructorID).Scan(&exists)

	if exists {
		return fmt.Errorf("giảng viên đã có trong danh sách dạy lớp này")
	}

	_, err = r.DB.Exec("INSERT INTO COURSE_INSTRUCTOR (course_id, instructor_id) VALUES (@p1, @p2)",
		courseID, instructorID)
	return err
}

func (r *Repository) GetStudentsByCourse(courseID int) ([]gin.H, error) {
	query := `
        SELECT s.id, s.name, e.mark 
        FROM STUDENT s
        JOIN ENROLLMENT e ON s.id = e.student_id
        WHERE e.course_id = @p1`

	rows, err := r.DB.Query(query, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []gin.H
	for rows.Next() {
		var id int
		var name string
		var mark float64
		if err := rows.Scan(&id, &name, &mark); err != nil {
			continue
		}
		students = append(students, gin.H{
			"id":   id,
			"name": name,
			"mark": mark,
		})
	}
	return students, nil
}

func (r *Repository) CallUpdateMarkSafe(ctx context.Context, sID, cID int, mark float64) error {
	_, err := r.DB.ExecContext(ctx, "EXEC UpdateMarkSafe @p1, @p2, @p3", sID, cID, mark)
	return err
}
func (r *Repository) CallUpdateMarkUnsafe(ctx context.Context, sID, cID int, mark float64) error {
	_, err := r.DB.ExecContext(ctx, "EXEC UpdateMarkUnsafe @p1, @p2, @p3", sID, cID, mark)
	return err
}

func (r *Repository) UpdateMarkSecure(sid, cid int, mark float64) error {

	query := `
        SET LOCK_TIMEOUT 50;
        UPDATE ENROLLMENT SET mark = @p1 WHERE student_id = @p2 AND course_id = @p3;
    `
	_, err := r.DB.Exec(query, mark, sid, cid)

	if err != nil {
		if isLockTimeoutError(err) {
			return errors.New("Hệ thống đang bận: Có giảng viên khác đang chốt báo cáo lớp này. Vui lòng đợi trong giây lát!")
		}
		return err
	}
	return nil
}

func isLockTimeoutError(err error) bool {
	if sqlErr, ok := err.(interface{ Number() int }); ok {
		return sqlErr.Number() == 1222
	}
	return false
}

func (r *Repository) CallFinalize(courseID int, isSafe bool, procedureName string) ([]gin.H, []gin.H, error) {
	rows, err := r.DB.Query("EXEC "+procedureName+" @p1, @p2", courseID, isSafe)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	phase1Data, err := scanReportRows(rows)
	if err != nil {
		return nil, nil, err
	}

	if !rows.NextResultSet() {
		return phase1Data, nil, fmt.Errorf("không tìm thấy dữ liệu phase 2")
	}

	phase2Data, err := scanReportRows(rows)
	if err != nil {
		return phase1Data, nil, err
	}

	return phase1Data, phase2Data, nil
}

func scanReportRows(rows *sql.Rows) ([]gin.H, error) {
	var data []gin.H
	for rows.Next() {
		var sid int
		var mark float64
		if err := rows.Scan(&sid, &mark); err != nil {
			return nil, err
		}
		data = append(data, gin.H{"sid": sid, "mark": mark})
	}
	return data, nil
}
