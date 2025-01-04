package hidden_crawler

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

type FormFiller struct{}

// Input tipine uygun değerler döndüren fonksiyon
func (ff *FormFiller) GetValue(inputType, placeholder, pattern string, maxLength int) string {
	defaultValue := "Test Value" // Varsayılan değer

	switch inputType {
	case "email":
		defaultValue = "test@example.com"
	case "tel":
		defaultValue = "+1234567890"
	case "password":
		defaultValue = "Password123!"
	case "number":
		defaultValue = "12345"
	case "url":
		defaultValue = "https://example.com"
	case "textarea":
		defaultValue = "This is a sample text for textarea."
	default: // Varsayılan olarak text tipi için
		defaultValue = "Test Value"
	}

	if defaultValue == "Test Value" {
		switch strings.ToLower(placeholder) {
		case "email":
			defaultValue = "test@example.com"
		case "tel":
			defaultValue = "+1234567890"
		case "password":
			defaultValue = "Password123!"
		case "number":
			defaultValue = "12345"
		case "url":
			defaultValue = "https://example.com"
		case "textarea":
			defaultValue = "This is a sample text for textarea."
		default: // Varsayılan olarak text tipi için
			defaultValue = "Test Value"
		}
	}

	// Eğer bir regex pattern'i varsa, uygun bir örnek değer döndür
	if pattern != "" {
		// Örnek regex oluşturma işlemi: Bu aşamada bir regex kütüphanesi kullanılarak uygun değer üretilebilir.
		matched, _ := regexp.MatchString(pattern, defaultValue)
		if !matched {
			defaultValue = "regex-compatible" // Regex'e uyumlu bir varsayılan değer
		}
	}

	// Eğer maxlength kısıtlaması varsa, değeri buna göre kısalt
	if maxLength > 0 && len(defaultValue) > maxLength {
		defaultValue = defaultValue[:maxLength]
	}

	return defaultValue
}

func (ff *FormFiller) ProcessInput(ctx context.Context) ([]map[string]interface{}, error) {
	var formInputDetails []map[string]interface{}
	// Tüm input bilgilerini al
	err := chromedp.Run(ctx,
		chromedp.EvaluateAsDevTools(`
		Array.from(document.querySelectorAll('form')).map(form => ({
            action: form.action || "",
            inputs: Array.from(form.querySelectorAll('input, textarea, select, button'))
                .filter(input => input.type !== "hidden") // Gizli alanları hariç tut
                .map(input => ({
                    tagName: input.tagName.toLowerCase(), // Tag adını küçük harf yap
                    placeholder: input.placeholder || "",
                    id: input.id || "",
                    name: input.name || "",
                    type: input.type || "text",
                    pattern: input.pattern || "",
                    maxLength: input.maxLength > 0 ? input.maxLength : -1
                }))
        }))
		`, &formInputDetails),
	)
	return formInputDetails, err
}

func (ff *FormFiller) GetaHref(ctx context.Context) ([]string, error) {
	var htmlLinks []string
	// Tüm input bilgilerini al
	err := chromedp.Run(ctx,
		chromedp.EvaluateAsDevTools(`
		(() => {
			const urls = [];
			document.querySelectorAll('a[href]').forEach(el => {
				if (el.href) urls.push(el.href);
				if (el.src) urls.push(el.src);
			});
			return urls;
		})()
			`, &htmlLinks),
	)

	return htmlLinks, err
}

func (ff *FormFiller) SendForms(ctx context.Context, hlClient *headlessClient, target string) {
	formInputDetails, err := ff.ProcessInput(ctx)
	if err != nil {
		//log.Fatalf("Failed to process input: %v", err)
	}
	// Her form için işlemleri yap
	for formIndex, form := range formInputDetails {
		action, ok_ := form["action"].(string)
		if !ok_ {
			continue
		}
		if action == "" || action == "/" {
			action = target + action
		}
		if !hlClient.isFormUniqe(action) {
			continue
		}
		var tasks = getChromedpNavigateTask(target)
		var submitBtnSelector string
		// Form içindeki tüm inputları doldur
		inputs, ok := form["inputs"].([]interface{}) // form["inputs"] bir slice olacak, type assertion yapıyoruz
		if !ok {
			continue
		}
		for _, inputmap := range inputs {

			input, ok := inputmap.(map[string]interface{}) // her input da bir map olmalı
			if !ok {
				continue
			}

			// Input adını ve tipini al
			tagName := input["tagName"].(string)
			inputId := input["id"].(string)
			inputName := input["name"].(string)
			placeholder := input["placeholder"].(string)
			inputType := input["type"].(string)
			pattern := input["pattern"].(string)
			maxLength := int(input["maxLength"].(float64))

			if strings.ToLower(inputType) == "hidden" {
				continue
			}

			if strings.ToLower(tagName) == "button" && strings.ToLower(inputType) == "submit" {
				submitBtnSelector = fmt.Sprintf(`form:nth-of-type(%d) button[type="submit"]`, formIndex+1)
				continue
			}
			if strings.ToLower(tagName) == "input" && strings.ToLower(inputType) == "submit" {
				submitBtnSelector = fmt.Sprintf(`form:nth-of-type(%d) input[type="submit"]`, formIndex+1)
				continue
			}

			// Input tipine göre uygun değer al
			value := ff.GetValue(inputType, placeholder, pattern, maxLength)
			// Seçici input veya textarea'ya göre ayarlanır
			var selector string
			if inputId != "" && inputName != "" {
				// Hem id hem de name varsa her ikisini de kullan
				selector = fmt.Sprintf(`%s[name="%s"], %s[id="%s"]`, tagName, inputName, tagName, inputId)
			} else if inputId != "" {
				// Sadece id varsa id'yi kullan
				selector = fmt.Sprintf(`%s[id="%s"]`, tagName, inputId)
			} else if inputName != "" {
				// Sadece name varsa name'i kullan
				selector = fmt.Sprintf(`%s[name="%s"]`, tagName, inputName)
			} else {
				// Hem id hem de name yoksa, tagName ile her öğeyi seç
				selector = fmt.Sprintf(`%s`, tagName)
			}

			// Görev listesine SendKeys ekle
			//var isElementExist bool
			//tasks = append(tasks, chromedp.Evaluate(fmt.Sprintf(`document.querySelector('%s') !== null`, selector), &isElementExist))
			tasks = append(tasks, logTask(selector+" - "+value, chromedp.SendKeys(selector, value)))

		}
		if submitBtnSelector != "" {
			tasks = append(tasks, logTask(submitBtnSelector, chromedp.Click(submitBtnSelector)))
		} else {
			// Formu gönder
			//tasks = append(tasks, logTask("form:nth-of-type", chromedp.Submit(fmt.Sprintf(`form:nth-of-type(%d)`, formIndex+1))))

		}
		tasks = append(tasks, chromedp.Sleep(1*time.Second))

		// Timeout ayarlamak için context ile birlikte kullan
		timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second) // 30 saniye timeout
		defer cancel()

		// İşlemleri çalıştır
		err = chromedp.Run(timeoutCtx, tasks)
		if err != nil {
			log.Printf("Error while submitting form %d: %v / target: ", formIndex+1, err, target)
			continue
		}

		//fmt.Printf("Form %d başarıyla gönderildi.\n", formIndex+1)
	}
}

func logTask(name string, task chromedp.Action) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		//log.Printf("Starting task: %s", name)
		err := task.Do(ctx)
		if err != nil {
			//log.Printf("Task failed: %s, error: %v", name, err)
			return err
		}
		//log.Printf("Task completed: %s", name)
		return nil
	})
}
