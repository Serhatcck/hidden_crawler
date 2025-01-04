(() => {
    console.log("Form analiz ve otomasyon başlıyor...");

    // Kullanılacak örnek değerler
    const defaultValues = {
        text: "sample text",
        email: "example@example.com",
        password: "Password123!",
        number: 42,
        url: "https://example.com",
        tel: "123-456-7890",
        date: "2024-12-31",
        datetime: "2024-12-31T23:59",
        datetimeLocal: "2024-12-31T23:59",
        month: "2024-12",
        week: "2024-W52",
        color: "#ff5733",
        checkbox: true, // İşaretle
        radio: true, // İlk radio seçeneği seçilir
    };

    // Tüm formları al
    const forms = document.forms;

    Array.from(forms).forEach((form, formIndex) => {

        // Formun tüm input elemanlarını al
        const inputs = form.querySelectorAll("input, textarea, select");

        inputs.forEach((input, inputIndex) => {
            try {
                const type = input.type || "text"; // Default olarak "text" kabul ediliyor
                // Değeri belirle ve input'u doldur
                if (type in defaultValues) {
                    if (type === "checkbox" || type === "radio") {
                        input.checked = defaultValues[type]; // Checkbox veya radio işaretle
                    } else {
                        input.value = defaultValues[type]; // Diğer input türleri için value ata
                    }
                } else {
                    // Bilinmeyen türler için log
                    
                }
            } catch (error) {
                }
        });

        // Formu gönder
        try {
            form.submit();
        } catch (error) {
        }
    });

    console.log("Form analiz ve otomasyon tamamlandı.");
})();