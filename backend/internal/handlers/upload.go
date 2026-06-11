package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// UploadPhoto обробляє завантаження фотографій (наприклад, фото паркування)
func UploadPhoto() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Обмежуємо максимальний розмір файлу до 5 МБ
		err := r.ParseMultipartForm(5 << 20)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Файл занадто великий (максимум 5 МБ)")
			return
		}

		// 2. Отримуємо файл з поля "photo" (так його має надсилати React)
		file, handler, err := r.FormFile("photo")
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Помилка завантаження файлу. Переконайтесь, що поле називається 'photo'")
			return
		}
		defer file.Close()

		// 2.2 Сувора перевірка розширення файлу в назві
		ext := strings.ToLower(filepath.Ext(handler.Filename))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
			RespondWithError(w, http.StatusBadRequest, "Недопустиме розширення файлу. Дозволені лише: .jpg, .jpeg, .png, .webp")
			return
		}

		// 2.5 Перевіряємо реальний тип файлу (MIME type)
		buff := make([]byte, 512)
		if _, err := file.Read(buff); err != nil && err != io.EOF {
			RespondWithError(w, http.StatusInternalServerError, "Помилка перевірки файлу")
			return
		}

		fileType := http.DetectContentType(buff)
		if fileType != "image/jpeg" && fileType != "image/png" && fileType != "image/webp" {
			RespondWithError(w, http.StatusBadRequest, "Дозволено завантажувати лише зображення (JPG, PNG, WEBP)")
			return
		}

		// Повертаємо вказівник читання файлу на самий початок, щоб скопіювати його цілком
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Помилка обробки файлу")
			return
		}

		// 3. Створюємо папку uploads, якщо вона ще не існує
		uploadDir := "uploads"
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Помилка сервера при збереженні файлу")
			return
		}

		// 4. Генеруємо унікальне ім'я файлу, щоб не перезаписати існуючі (наприклад: 167890123_image.jpg)
		fileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(handler.Filename))
		filePath := filepath.Join(uploadDir, fileName)

		// 5. Створюємо порожній файл на диску
		dst, err := os.Create(filePath)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Не вдалося створити файл на сервері")
			return
		}
		defer dst.Close()

		// 6. Копіюємо вміст завантаженого файлу у створений файл на диску
		if _, err := io.Copy(dst, file); err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Помилка запису файлу")
			return
		}

		// 7. Повертаємо клієнту готовий URL, який він потім передасть в EndRide
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"url": fmt.Sprintf("/uploads/%s", fileName),
		})
	}
}
