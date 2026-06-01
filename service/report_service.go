package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"platform-intern-growth/models"
	"platform-intern-growth/repository"

	"github.com/google/uuid"
)

type TaskSummary struct {
	Title       string   `json:"title"`
	Status      string   `json:"status"`
	Competences []string `json:"competences"`
}

type StudentReport struct {
	StudentID          string                        `json:"student_id"`
	StudentName        string                        `json:"student_name"`
	Period             string                        `json:"period"`
	CompletedTasks     int                           `json:"completed_tasks"`
	TotalTasks         int                           `json:"total_tasks"`
	CompetencesCovered int                           `json:"competences_covered"`
	Tasks              []TaskSummary                 `json:"tasks"`
	Competences        []models.CompetenceMatrixItem `json:"competences"`
}

func GenerateStudentReport(studentID uuid.UUID) (*StudentReport, error) {
	student, err := repository.FindUserByID(studentID)
	if err != nil {
		return nil, fmt.Errorf("стажёр не найден")
	}
	if student.Role != models.RoleIntern {
		return nil, fmt.Errorf("указанный пользователь не является стажёром")
	}

	allTasks, err := repository.FindTasksByStudentID(studentID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить задачи стажёра")
	}

	matrix, err := GetStudentMatrix(studentID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить матрицу компетенций")
	}

	completedCount := 0
	var taskSummaries []TaskSummary
	for _, task := range allTasks {
		if task.Status == models.StatusCompleted {
			completedCount++
		}
		taskSummaries = append(taskSummaries, TaskSummary{
			Title:       task.Title,
			Status:      string(task.Status),
			Competences: []string(task.Competences),
		})
	}

	coveredCount := 0
	for _, item := range matrix.Competences {
		if item.Status == models.CompetenceStatusCovered {
			coveredCount++
		}
	}

	now := time.Now()
	period := fmt.Sprintf("%s %d", russianMonth(now.Month()), now.Year())

	return &StudentReport{
		StudentID:          studentID.String(),
		StudentName:        student.Name,
		Period:             period,
		CompletedTasks:     completedCount,
		TotalTasks:         len(allTasks),
		CompetencesCovered: coveredCount,
		Tasks:              taskSummaries,
		Competences:        matrix.Competences,
	}, nil
}

// GenerateStudentReportJSON возвращает отчёт как красиво отформатированный JSON-файл.
func GenerateStudentReportJSON(studentID uuid.UUID) ([]byte, error) {
	report, err := GenerateStudentReport(studentID)
	if err != nil {
		return nil, err
	}

	payload := map[string]interface{}{
		"success": true,
		"data":    report,
	}

	jsonBytes, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("не удалось сериализовать отчёт: %w", err)
	}
	return jsonBytes, nil
}

// GenerateStudentReportPDF возвращает PDF-файл с отчётом стажёра.
// Генерируется напрямую в байты без сторонних библиотек.
func GenerateStudentReportPDF(studentID uuid.UUID) ([]byte, error) {
	report, err := GenerateStudentReport(studentID)
	if err != nil {
		return nil, err
	}

	// Собираем строки отчёта (только ASCII через транслитерацию)
	var lines []string
	lines = append(lines, "STUDENT PROGRESS REPORT")
	lines = append(lines, strings.Repeat("=", 55))
	lines = append(lines, fmt.Sprintf("Name:    %s", transliterate(report.StudentName)))
	lines = append(lines, fmt.Sprintf("Period:  %s", transliterate(report.Period)))
	lines = append(lines, fmt.Sprintf("Tasks:   %d total, %d completed", report.TotalTasks, report.CompletedTasks))
	lines = append(lines, fmt.Sprintf("Competences covered: %d", report.CompetencesCovered))
	lines = append(lines, "")
	lines = append(lines, "TASKS")
	lines = append(lines, strings.Repeat("-", 55))
	if len(report.Tasks) == 0 {
		lines = append(lines, "(no tasks yet)")
	}
	for _, task := range report.Tasks {
		lines = append(lines, fmt.Sprintf("[%-11s] %s", task.Status, transliterate(task.Title)))
	}
	lines = append(lines, "")
	lines = append(lines, "COMPETENCE MATRIX")
	lines = append(lines, strings.Repeat("-", 55))
	if len(report.Competences) == 0 {
		lines = append(lines, "(no competences yet)")
	}
	for _, comp := range report.Competences {
		lines = append(lines, fmt.Sprintf("%-35s %s", transliterate(comp.Name), string(comp.Status)))
	}

	return buildRawPDF(lines), nil
}

// buildRawPDF создаёт минимальный валидный PDF 1.4 без сторонних библиотек.
// Использует встроенный шрифт Helvetica (ASCII only).
func buildRawPDF(textLines []string) []byte {
	// ── Шаг 1: контентный поток страницы ──────────────────────────────────
	var streamBuf bytes.Buffer
	streamBuf.WriteString("BT\n")
	streamBuf.WriteString("/F1 11 Tf\n") // шрифт Helvetica, 11pt
	streamBuf.WriteString("14 TL\n")    // межстрочный интервал 14pt
	streamBuf.WriteString("50 800 Td\n") // начало текста: x=50, y=800 (от нижнего края A4)
	for _, line := range textLines {
		// Экранируем символы, зарезервированные в PDF
		line = strings.ReplaceAll(line, "\\", "\\\\")
		line = strings.ReplaceAll(line, "(", "\\(")
		line = strings.ReplaceAll(line, ")", "\\)")
		streamBuf.WriteString("(" + line + ") Tj T*\n")
	}
	streamBuf.WriteString("ET\n")
	streamBytes := streamBuf.Bytes()

	// ── Шаг 2: строим тело PDF в bytes.Buffer, отслеживаем байтовые смещения ──
	var body bytes.Buffer
	var offsets [6]int

	body.WriteString("%PDF-1.4\n")

	offsets[1] = body.Len()
	body.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")

	offsets[2] = body.Len()
	body.WriteString("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")

	offsets[3] = body.Len()
	body.WriteString("3 0 obj\n")
	body.WriteString("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842]\n")
	body.WriteString("   /Contents 4 0 R\n")
	body.WriteString("   /Resources << /Font << /F1 5 0 R >> >> >>\n")
	body.WriteString("endobj\n")

	// Объект 4: поток содержимого. Length = ровно len(streamBytes).
	offsets[4] = body.Len()
	body.WriteString(fmt.Sprintf("4 0 obj\n<< /Length %d >>\nstream\n", len(streamBytes)))
	body.Write(streamBytes) // streamBytes заканчивается на "ET\n"
	body.WriteString("endstream\nendobj\n")

	// Объект 5: шрифт (встроен в любой PDF viewer)
	offsets[5] = body.Len()
	body.WriteString("5 0 obj\n")
	body.WriteString("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica\n")
	body.WriteString("   /Encoding /WinAnsiEncoding >>\n")
	body.WriteString("endobj\n")

	// ── Шаг 3: таблица xref — каждая запись РОВНО 20 байт ─────────────────
	xrefOffset := body.Len()
	body.WriteString("xref\n")
	body.WriteString("0 6\n")
	body.WriteString("0000000000 65535 f \n") // 20 байт: свободная запись
	for i := 1; i <= 5; i++ {
		body.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[i])) // 20 байт
	}

	// ── Шаг 4: трейлер ─────────────────────────────────────────────────────
	body.WriteString("trailer\n<< /Size 6 /Root 1 0 R >>\n")
	body.WriteString(fmt.Sprintf("startxref\n%d\n", xrefOffset))
	body.WriteString("%%EOF\n")

	return body.Bytes()
}

func russianMonth(month time.Month) string {
	months := map[time.Month]string{
		time.January: "yanvar", time.February: "fevral", time.March: "mart",
		time.April: "aprel", time.May: "may", time.June: "iyun",
		time.July: "iyul", time.August: "avgust", time.September: "sentyabr",
		time.October: "oktyabr", time.November: "noyabr", time.December: "dekabr",
	}
	return months[month]
}

func transliterate(text string) string {
	translitMap := map[rune]string{
		'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d",
		'е': "e", 'ё': "yo", 'ж': "zh", 'з': "z", 'и': "i",
		'й': "j", 'к': "k", 'л': "l", 'м': "m", 'н': "n",
		'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t",
		'у': "u", 'ф': "f", 'х': "kh", 'ц': "ts", 'ч': "ch",
		'ш': "sh", 'щ': "shch", 'ъ': "", 'ы': "y", 'ь': "",
		'э': "e", 'ю': "yu", 'я': "ya",
		'А': "A", 'Б': "B", 'В': "V", 'Г': "G", 'Д': "D",
		'Е': "E", 'Ё': "Yo", 'Ж': "Zh", 'З': "Z", 'И': "I",
		'Й': "J", 'К': "K", 'Л': "L", 'М': "M", 'Н': "N",
		'О': "O", 'П': "P", 'Р': "R", 'С': "S", 'Т': "T",
		'У': "U", 'Ф': "F", 'Х': "Kh", 'Ц': "Ts", 'Ч': "Ch",
		'Ш': "Sh", 'Щ': "Shch", 'Ъ': "", 'Ы': "Y", 'Ь': "",
		'Э': "E", 'Ю': "Yu", 'Я': "Ya",
	}
	result := make([]byte, 0, len(text))
	for _, char := range text {
		if latin, found := translitMap[char]; found {
			result = append(result, []byte(latin)...)
		} else {
			result = append(result, []byte(string(char))...)
		}
	}
	return string(result)
}
