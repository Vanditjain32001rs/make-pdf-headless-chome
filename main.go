package main

import (
	"bytes"
	"context"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
)

type Data struct {
	Title   string
	Content string
}

func main() {
	router := gin.Default()

	router.POST("/pdf", func(c *gin.Context) {
		data := generateRandomData()

		htmlContent, err := renderHTML(data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		pdfData, err := generatePDF(htmlContent)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", "attachment; filename=output.pdf")
		c.Data(http.StatusOK, "application/pdf", pdfData)
	})

	//router.LoadHTMLGlob("templates/*")

	log.Println("Starting server on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

func generateRandomData() Data {
	rand.Seed(time.Now().UnixNano())
	titles := []string{"Random Title 1", "Random Title 2", "Random Title 3"}
	contents := []string{
		"This is some random content to be converted to PDF.",
		"Here is another set of random content for the PDF generation.",
		"Yet another random content example for PDF conversion.",
	}

	return Data{
		Title:   titles[rand.Intn(len(titles))],
		Content: contents[rand.Intn(len(contents))],
	}
}

func renderHTML(data Data) (string, error) {
	tmpl := `
    <!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <title>{{ .Title }}</title>
    </head>
    <body>
        <h1>{{ .Title }}</h1>
        <p>{{ .Content }}</p>
    </body>
    </html>
    `
	t, err := template.New("webpage").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func generatePDF(htmlContent string) ([]byte, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var pdfData []byte
	err := chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate("data:text/html," + htmlContent),
		chromedp.Sleep(2 * time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().WithPrintBackground(true).Do(ctx)
			pdfData = buf
			return err
		}),
	})

	return pdfData, err
}
