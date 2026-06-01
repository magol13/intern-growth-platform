package handlers

import (
	"net/http"

	"platform-intern-growth/middleware"
	"platform-intern-growth/models"
	"platform-intern-growth/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// respondWithReport — вспомогательная функция: отдаёт отчёт в нужном формате.
// format=json → скачивается файл report.json
// format=pdf  → скачивается файл report.pdf
func respondWithReport(context *gin.Context, studentID uuid.UUID, format string) {
	if format == "pdf" {
		pdfBytes, err := service.GenerateStudentReportPDF(studentID)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Не удалось сгенерировать PDF: " + err.Error(),
			})
			return
		}
		context.Header("Content-Disposition", "attachment; filename=report.pdf")
		context.Data(http.StatusOK, "application/pdf", pdfBytes)
		return
	}

	// По умолчанию — JSON как скачиваемый файл
	jsonBytes, err := service.GenerateStudentReportJSON(studentID)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	context.Header("Content-Disposition", "attachment; filename=report.json")
	context.Data(http.StatusOK, "application/json", jsonBytes)
}

// GetMentorStudentReport — ментор скачивает отчёт по конкретному стажёру.
// GET /students/:id/report?format=json|pdf
func GetMentorStudentReport(context *gin.Context) {
	_ = context.MustGet(middleware.CurrentUserKey).(*models.User)

	studentIDString := context.Param("id")
	studentID, err := uuid.Parse(studentIDString)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Неверный формат ID стажёра",
		})
		return
	}

	format := context.DefaultQuery("format", "json")
	respondWithReport(context, studentID, format)
}

// GetMyReport — стажёр скачивает собственный отчёт.
// GET /my-report?format=json|pdf
func GetMyReport(context *gin.Context) {
	currentUser := context.MustGet(middleware.CurrentUserKey).(*models.User)
	format := context.DefaultQuery("format", "json")
	respondWithReport(context, currentUser.ID, format)
}
