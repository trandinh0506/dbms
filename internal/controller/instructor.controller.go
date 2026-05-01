package controller

import (
	"dbms/internal/common"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (ctrl *Controller) GetMyCourses(c *gin.Context) {
	instructorID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy ID giảng viên"})
		return
	}

	courses, err := ctrl.Svc.Repo.GetCoursesByInstructor(instructorID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi lấy danh sách lớp: " + err.Error()})
		return
	}

	if len(courses) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "Bạn chưa được phân công dạy lớp nào", "data": []string{}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"instructor_id": instructorID,
		"courses":       courses,
	})
}
func (ctrl *Controller) UpdateStudentMark(c *gin.Context) {
	role := c.MustGet("role").(string)
	if role != "Instructor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Chỉ giảng viên mới có quyền sửa điểm"})
		return
	}

	mode := c.DefaultQuery("mode", "safe")
	sID, _ := strconv.Atoi(c.Query("sid"))
	cID, _ := strconv.Atoi(c.Query("cid"))
	mark, _ := strconv.ParseFloat(c.Query("mark"), 64)

	var err error
	if mode == "safe" || mode == "unsafe" {
		err = ctrl.Svc.UpdateMark(sID, cID, mark, mode)
	} else {
		err = ctrl.Svc.UpdateMarkBuggy(sID, cID, mark)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "Cập nhật thành công", "mode": mode})
}

func (ctrl *Controller) CreateCourse(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "Instructor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Chỉ giảng viên mới có quyền tạo khóa học"})
		return
	}

	instructorID := c.MustGet("user_id").(int)

	var req struct {
		Name     string `json:"name" binding:"required"`
		Capacity int    `json:"capacity"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tên khóa học không được để trống"})
		return
	}

	if req.Capacity <= 0 {
		req.Capacity = common.DEFAULT_STUDENTS_PER_COURSE
	}

	courseID, err := ctrl.Svc.Repo.CreateCourse(req.Name, req.Capacity, instructorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tạo khóa học: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Tạo khóa học thành công",
		"course_id": courseID,
		"name":      req.Name,
	})
}

func (ctrl *Controller) InviteInstructor(c *gin.Context) {
	currentUserID := c.MustGet("user_id").(int)

	var req struct {
		CourseID     int `json:"course_id" binding:"required"`
		InstructorID int `json:"instructor_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Thiếu thông tin khóa học hoặc giảng viên được mời"})
		return
	}

	var ownerID int
	err := ctrl.Svc.Repo.DB.QueryRow("SELECT created_by FROM COURSE WHERE id = @p1", req.CourseID).Scan(&ownerID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Không tìm thấy khóa học"})
		return
	}

	if ownerID != currentUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Chỉ giảng viên tạo khóa học mới có quyền mời người khác"})
		return
	}

	err = ctrl.Svc.Repo.InviteInstructor(req.CourseID, req.InstructorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Đã mời giảng viên tham gia dạy thành công!"})
}

func (ctrl *Controller) GetStudentsByCourse(c *gin.Context) {
	courseID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID khóa học không hợp lệ"})
		return
	}

	students, err := ctrl.Svc.GetCourseStudents(courseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"course_id": courseID,
		"students":  students,
	})
}

func (ctrl *Controller) StartSession(c *gin.Context) {

	session_id, err := ctrl.Svc.Ssm.StartSession(ctrl.Svc.Repo.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"session_id": session_id})
}

func (ctrl *Controller) UpdateMarkInSession(c *gin.Context) {
	var req struct {
		SessionID string  `json:"session_id" binding:"required"`
		SID       int     `json:"sid" binding:"required"`
		CID       int     `json:"cid" binding:"required"`
		Mark      float64 `json:"mark" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu thông tin"})
		return
	}

	err := ctrl.Svc.Ssm.UpdateDraft(req.SessionID, req.SID, req.CID, req.Mark)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Cập nhật thành công"})
}

func (ctrl *Controller) CommitSession(c *gin.Context) {
	sessionID := c.Query("session_id")
	err := ctrl.Svc.Ssm.CloseSession(sessionID, "COMMIT")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "Commit thành công"})
}

func (ctrl *Controller) RollbackSession(c *gin.Context) {
	sessionID := c.Query("session_id")
	err := ctrl.Svc.Ssm.CloseSession(sessionID, "ROLLBACK")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "Rollback thành công"})
}

func (ctrl *Controller) FinalizeReport(c *gin.Context) {
	var req struct {
		CourseID int  `json:"course_id" binding:"required"`
		IsSafe   bool `json:"is_safe"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu thông tin"})
		return
	}

	message, reportData, err := ctrl.Svc.ProcessFinalReport(req.CourseID, req.IsSafe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": message,
		"report":  reportData,
	})

}

func (ctrl *Controller) GetWeakStudentReport(c *gin.Context) {
	var req struct {
		CourseID int  `json:"course_id" binding:"required"`
		IsSafe   bool `json:"is_safe"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu thông tin"})
		return
	}

	message, reportData, err := ctrl.Svc.ProcessWeakStudentReport(req.CourseID, req.IsSafe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": message,
		"report":  reportData,
	})
}
