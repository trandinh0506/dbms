package router

import (
	"dbms/internal/controller"
	"dbms/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(ctrl *controller.Controller) *gin.Engine {
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	// Public routes
	r.POST("/api/login", ctrl.Login)
	r.POST("/api/register", ctrl.Register)

	// Private routes
	auth := r.Group("/api")
	auth.Use(middleware.AuthMiddleware(ctrl.Config))
	{
		// Common routes
		auth.GET("/courses", ctrl.GetAllCourses)

		// Student routes
		student := auth.Group("/student")
		{
			student.POST("/enroll", ctrl.Enroll)
			student.GET("/grades-unsafe", ctrl.GetGradesUnsafe)
			student.GET("/grades-safe", ctrl.GetGradesSafe)
		}

		// Instructor routes
		instructor := auth.Group("/instructor")
		instructor.Use(middleware.RoleCheckMiddleware("Instructor"))
		{

			instructor.GET("/courses", ctrl.GetMyCourses)
			instructor.POST("/courses", ctrl.CreateCourse)
			instructor.POST("/invite", ctrl.InviteInstructor)
			instructor.GET("/course-students/:id", ctrl.GetStudentsByCourse)
			instructor.POST("/start-session", ctrl.StartSession)
			instructor.PUT("/update-mark-in-session", ctrl.UpdateMarkInSession)
			instructor.POST("/commit", ctrl.CommitSession)
			instructor.POST("/rollback", ctrl.RollbackSession)
			instructor.POST("/final-report", ctrl.FinalizeReport)
			instructor.POST("/weak-report", ctrl.GetWeakStudentReport)
		}
	}

	return r
}
