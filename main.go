package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/Michael-Wilburn/car_admin_panel/db"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
)

type Car struct {
	ID         uuid.UUID
	Online     bool
	TypeCar    string
	Brand      string
	Model      string
	Year       int
	Kilometers int64
	CarDomain  string
	Price      float64
	InfoPrice  float64
	Currency   string
}

type PageData struct {
	Cars  []Car
	Count int
}

type FormData struct {
	Car   Car
	Count int
}

var DB *sql.DB

func main() {
	// Establecer conexión a la base de datos
	database, err := db.EstablishDbConnection()
	if err != nil {
		log.Fatal("Error establishing database connection:", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Println("Error closing database connection:", err)
		}
	}()

	DB = database

	routes := mux.NewRouter()
	routes.HandleFunc("/", ServeIndex).Methods("GET")
	routes.HandleFunc("/api/file/excel", DownloadExcel)
	routes.HandleFunc("/api/file/pdf", DownloadPDF)

	routes.HandleFunc("/api/car", CreateCar).Methods("POST")
	routes.HandleFunc("/api/car/{id}", GetCar).Methods("GET")
	routes.HandleFunc("/api/car/{id}", UpdateCar).Methods("PUT")
	routes.HandleFunc("/api/car/{id}", DeleteCar).Methods("DELETE")

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.Handle("/", routes)

	log.Print("Listening on :3000...")
	err = http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func ServeIndex(w http.ResponseWriter, r *http.Request) {
	var Cars = []Car{}
	cars, err := DB.Query("SELECT * FROM car ORDER BY brand ASC;")
	if err != nil {
		log.Fatal(err)
	}
	defer cars.Close()

	for cars.Next() {
		var thisCar Car
		// Escanear todos los campos de la fila actual y asignarlos al struct Car
		err = cars.Scan(&thisCar.ID, &thisCar.Online, &thisCar.TypeCar, &thisCar.Brand, &thisCar.Model, &thisCar.Year, &thisCar.Kilometers, &thisCar.CarDomain, &thisCar.Price, &thisCar.InfoPrice, &thisCar.Currency)
		if err != nil {
			log.Fatal(err)
		}

		// Agregar este Car a la lista de Cars
		Cars = append(Cars, thisCar)
	}
	Count := len(Cars)

	pageData := PageData{
		Cars:  Cars,
		Count: Count,
	}
	// Generar el archivo excel
	generateExcel(Cars)

	// Configurar las funciones del template antes de llamar a ParseFiles
	t := template.New("index.html").Funcs(template.FuncMap{"formatNumber": formatNumber})

	t, err = t.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatal(err)
	}

	err = t.Execute(w, pageData)
	if err != nil {
		log.Fatal(err)
	}
}

// Función para formatear un número a una cadena con separador de miles
func formatNumber(num float64) string {
	// Convertir el número a una cadena sin decimales
	numStr := strconv.FormatFloat(num, 'f', 0, 64)

	// Dividir la cadena en partes de tres dígitos desde la derecha
	parts := make([]string, 0)
	for len(numStr) > 3 {
		parts = append([]string{numStr[len(numStr)-3:]}, parts...)
		numStr = numStr[:len(numStr)-3]
	}
	parts = append([]string{numStr}, parts...)

	// Unir las partes con un punto como separador
	formatted := strings.Join(parts, ".")

	return formatted
}

func generateExcel(cars []Car) (*bytes.Buffer, error) {
	file := excelize.NewFile()

	// Create a new sheet
	_, err := file.NewSheet("Sheet1")
	if err != nil {
		return nil, fmt.Errorf("error creating new sheet: %v", err)
	}

	// Write column headers to the first row starting from column B
	headers := []string{"Marca", "Modelo", "Año", "Kilómetros", "Patente", "Precio"}
	for col, header := range headers {
		file.SetCellValue("Sheet1", fmt.Sprintf("%c%d", 'B'+col, 1), header)
	}

	style, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Family: "Futura",
			Size:   16,
			Color:  "000000",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("error creating new style: %v", err)
	}

	tableStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Futura",
			Size:   12,
			Color:  "000000",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("error creating new style: %v", err)
	}
	expPrice := `"$"#,##0`
	priceStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Futura",
			Size:   12,
			Color:  "000000",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		CustomNumFmt: &expPrice,
	})

	if err != nil {
		return nil, fmt.Errorf("error creating new style: %v", err)
	}
	expKilometers := "#,##0"
	kilometerStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Futura",
			Size:   12,
			Color:  "000000",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		CustomNumFmt: &expKilometers,
	})

	if err != nil {
		return nil, fmt.Errorf("error creating new style: %v", err)
	}

	// Write data to the Excel file starting from row 2
	for row, car := range cars {
		file.SetCellValue("Sheet1", fmt.Sprintf("B%d", row+2), car.Brand)
		file.SetCellValue("Sheet1", fmt.Sprintf("C%d", row+2), car.Model)
		file.SetCellValue("Sheet1", fmt.Sprintf("D%d", row+2), car.Year)
		file.SetCellValue("Sheet1", fmt.Sprintf("E%d", row+2), car.Kilometers)
		file.SetCellValue("Sheet1", fmt.Sprintf("F%d", row+2), car.CarDomain)
		file.SetCellValue("Sheet1", fmt.Sprintf("G%d", row+2), car.Price)
	}

	file.SetCellStyle("Sheet1", "B1", "G"+fmt.Sprint((len(cars)+1)), tableStyle)
	file.SetCellStyle("Sheet1", "G1", "G"+fmt.Sprint((len(cars)+1)), priceStyle)
	file.SetCellStyle("Sheet1", "E1", "E"+fmt.Sprint((len(cars)+1)), kilometerStyle)
	file.SetCellStyle("Sheet1", "B1", "G1", style)

	file.SetColWidth("Sheet1", "A", "H", 20)
	file.SetColWidth("Sheet1", "C", "C", 60)

	// Define la altura deseada para las celdas
	cellHeight := 20

	// Establece la altura de las celdas en todas las filas (desde la segunda fila hasta la última)
	for row := 1; row <= len(cars)+1; row++ {
		err := file.SetRowHeight("Sheet1", row, float64(cellHeight))
		if err != nil {
			return nil, fmt.Errorf("error setting row height: %v", err)
		}
	}

	// Set automatic filter on the first row
	options := []excelize.AutoFilterOptions{}
	err = file.AutoFilter("Sheet1", "B1:G1", options)
	if err != nil {
		log.Println("Error setting automatic filter:", err)
	}

	// Write the Excel file to a buffer
	buf := new(bytes.Buffer)
	if err := file.Write(buf); err != nil {
		return nil, fmt.Errorf("error writing Excel file to buffer: %v", err)
	}

	return buf, nil
}

func DownloadExcel(w http.ResponseWriter, r *http.Request) {
	var Cars = []Car{}
	cars, err := DB.Query("SELECT * FROM car ORDER BY brand ASC;")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cars.Close()

	for cars.Next() {
		var thisCar Car
		err = cars.Scan(&thisCar.ID, &thisCar.Online, &thisCar.TypeCar, &thisCar.Brand, &thisCar.Model, &thisCar.Year, &thisCar.Kilometers, &thisCar.CarDomain, &thisCar.Price, &thisCar.InfoPrice, &thisCar.Currency)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		Cars = append(Cars, thisCar)
	}

	// Generar el archivo Excel
	excelData, err := generateExcel(Cars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Establecer los encabezados para la descarga del archivo
	w.Header().Set("Content-Disposition", "attachment; filename=cars.xlsx")
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")

	// Escribir los datos del archivo Excel en el cuerpo de la respuesta
	if _, err := w.Write(excelData.Bytes()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

///////////////////////////////////////////

func DownloadPDF(w http.ResponseWriter, r *http.Request) {
	// Consulta a la base de datos para obtener los datos necesarios
	rows, err := DB.Query("SELECT brand, model, year, kilometers, car_domain, price, currency FROM car ORDER BY brand ASC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Cargar la fuente desde un archivo TTF
	fontPath := "./roboto/Roboto-Regular.ttf" // Ajusta la ruta y el nombre del archivo TTF
	fontBytes, err := os.ReadFile(fontPath)
	if err != nil {
		http.Error(w, "Failed to load font: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Crear un nuevo PDF
	pdf := gofpdf.New("P", "mm", "Letter", "")
	pdf.AddPage()

	// Añadir la fuente al PDF
	pdf.AddUTF8FontFromBytes("Roboto", "", fontBytes)

	// Configurar el estilo de fuente y tamaño
	pdf.SetFont("Roboto", "", 8)

	// Crear una fila para los encabezados de la tabla
	pdf.SetFillColor(220, 220, 220)
	pdf.SetTextColor(0, 0, 0)
	for _, colStr := range []string{"Marca", "Modelo", "Año", "Kilometros", "Dominio", "Precio"} {
		if colStr == "Modelo" {
			pdf.CellFormat(75, 10, colStr, "1", 0, "C", true, 0, "")
		} else {
			pdf.CellFormat(25, 10, colStr, "1", 0, "C", true, 0, "")
		}
	}
	pdf.Ln(-1)

	// Iterar sobre los resultados de la consulta y agregarlos al PDF
	for rows.Next() {
		var brand, model, carDomain string
		var year int
		var kilometers int64
		var price float64
		var currency string

		err := rows.Scan(&brand, &model, &year, &kilometers, &carDomain, &price, &currency)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Crear una fila para los datos de cada coche
		pdf.CellFormat(25, 8, brand, "1", 0, "C", false, 0, "")
		pdf.CellFormat(75, 8, model, "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, 8, fmt.Sprintf("%d", year), "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, 8, formatNumber(float64(kilometers)), "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, 8, carDomain, "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, 8, CurrencyType(currency, price), "1", 0, "C", false, 0, "")
		pdf.Ln(-1)
	}

	// Guardar el PDF en la respuesta HTTP
	w.Header().Set("Content-Disposition", "attachment; filename=cars.pdf")
	w.Header().Set("Content-Type", "application/pdf")
	err = pdf.Output(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func CurrencyType(currency string, price float64) string {
	priceFormated := formatNumber(price)
	if currency == "$" {
		return "$ " + priceFormated
	} else {
		return "USD " + priceFormated
	}
}

//////////////////////////////////////////

func CreateCar(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println("Error parsing form: ", err.Error())
		return
	}

	typeCar := r.FormValue("typeCar")
	brand := r.FormValue("brand")
	model := r.FormValue("model")
	year := r.FormValue("year")
	kilometers := r.FormValue("kilometers")
	carDomain := r.FormValue("licensePlate")
	currency := r.FormValue("currency")
	price := r.FormValue("price")
	infoPrice := r.FormValue("infoPrice")

	_, err = DB.Exec("INSERT INTO car (online, car_type, brand, model, year, kilometers, car_domain, price, info_price, currency) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)", true, typeCar, brand, model, year, kilometers, carDomain, price, infoPrice, currency)
	if err != nil {
		log.Println("Error inserting new car record: ", err.Error())
		return
	}
	fmt.Println("car created")

	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

func DeleteCar(w http.ResponseWriter, r *http.Request) {
	// Obtiene las variables de la URL
	vars := mux.Vars(r)
	id := vars["id"]
	fmt.Println("ID del coche a eliminar: ", id)

	// Ejecuta la consulta DELETE con el ID
	_, err := DB.Exec("DELETE FROM car WHERE id = $1;", id)
	if err != nil {
		log.Println("Error al eliminar el coche: ", err.Error())
	}

	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

func GetCar(w http.ResponseWriter, r *http.Request) {

	var car Car
	// Obtiene las variables de la URL
	vars := mux.Vars(r)
	id := vars["id"]
	fmt.Println("ID del coche a eliminar: ", id)

	row := DB.QueryRow("SELECT id, online, car_type, brand, model, year, kilometers, car_domain, price, info_price, currency FROM car WHERE id = $1;", id)

	err := row.Scan(&car.ID, &car.Online, &car.TypeCar, &car.Brand, &car.Model, &car.Year, &car.Kilometers, &car.CarDomain, &car.Price, &car.InfoPrice, &car.Currency)
	if err != nil {
		log.Println("Error al obtener los datos del coche: ", err.Error())
		http.Error(w, "No se pudo obtener el coche", http.StatusInternalServerError)
		return
	}

	var count int
	err = DB.QueryRow("SELECT COUNT(*) FROM car").Scan(&count)
	if err != nil {
		log.Println("Error querying car count: ", err.Error())
		http.Error(w, "Error fetching car count", http.StatusInternalServerError)
		return
	}

	pageData := FormData{
		Car:   car,
		Count: count,
	}

	t, _ := template.ParseFiles("templates/form.html")
	t.Execute(w, pageData)
}

func UpdateCar(w http.ResponseWriter, r *http.Request) {
	// Obtiene las variables de la URL
	vars := mux.Vars(r)
	id := vars["id"]

	// Analiza el formulario entrante
	err := r.ParseForm()
	if err != nil {
		log.Println("Error parsing form: ", err.Error())
		http.Error(w, "Error parsing form", http.StatusInternalServerError)
		return
	}

	typeCar := r.FormValue("typeCar")
	brand := r.FormValue("brand")
	model := r.FormValue("model")
	year := r.FormValue("year")
	kilometers := r.FormValue("kilometers")
	carDomain := r.FormValue("licensePlate")
	currency := r.FormValue("currency")
	price := r.FormValue("price")
	infoPrice := r.FormValue("infoPrice")

	// Ejecuta la consulta UPDATE
	_, err = DB.Exec("UPDATE car SET car_type = $1, brand = $2, model = $3, year = $4, kilometers = $5, car_domain = $6, price = $7, info_price = $8, currency = $9 WHERE id = $10",
		typeCar, brand, model, year, kilometers, carDomain, price, infoPrice, currency, id)
	if err != nil {
		log.Println("Error updating car record: ", err.Error())
		http.Error(w, "Error updating car", http.StatusInternalServerError)
		return
	}

	fmt.Println("Car updated successfully")
	http.Redirect(w, r, "/home", http.StatusSeeOther)
}
