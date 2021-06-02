package main

import (
	"context"
	"encoding/json"
	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	maps "googlemaps.github.io/maps"
	"log"
	"os"
	"strings"
	"time"
)

/*
 Giderilmesi gereken hatalar
	ağrı ilinde district ile name değişmesi lazım

*/

type Pharmacy struct {
	PharmacyName        string `json:"pharmacyName"`
	PharmacyAddress     string `json:"pharmacyAddress"`
	PharmacyPhoneNumber string `json:"pharmacyPhoneNumber"`
	PharmacyProvince    string `json:"pharmacyProvince"`
	PharmacyDistrict    string `json:"pharmacyDistrict"`
	PharmacyLatLng      string `json:"pharmacyLatLng"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	province := [81]string{
		"adana",
		"adiyaman",
		"afyonkarahisar",
		"agri",
		"amasya",
		"ankara",
		"antalya",
		"artvin",
		"aydin",
		"balikesir",
		"bilecik",
		"bingol",
		"bitlis",
		"bolu",
		"burdur",
		"bursa",
		"canakkale",
		"cankiri",
		"corum",
		"denizli",
		//"diyarbakir",
		"edirne",
		"elazig",
		"erzincan",
		"erzurum",
		"eskisehir",
		"gaziantep",
		"giresun",
		"gumushane",
		"hakkari",
		"hatay",
		"isparta",
		"mersin",
		"istanbul",
		"izmir",
		"kars",
		"kastamonu",
		"kayseri",
		"kirklareli",
		"kirsehir",
		"kocaeli",
		"konya",
		"kutahya",
		"malatya",
		"manisa",
		"kahramanmaras",
		"mardin",
		"mugla",
		"mus",
		"nevsehir",
		"nigde",
		"ordu",
		"rize",
		"sakarya",
		"samsun",
		"siirt",
		"sinop",
		"sivas",
		"tekirdag",
		"tokat",
		"trabzon",
		"tunceli",
		"sanliurfa",
		"usak",
		"van",
		"yozgat",
		"zonguldak",
		"aksaray",
		"bayburt",
		"karaman",
		"kirikkale",
		"batman",
		"sirnak",
		"bartin",
		"ardahan",
		"igdir",
		"yalova",
		"karabuk",
		"kilis",
		"osmaniye",
		"duzce",
	}
	for i := 0; i < len(province); i++ {
		getFetchDataProvince(province[i])
	}
}

func connectAndWriteMongoDb(pharmacyData []byte) {
	var data []interface{}
	err := json.Unmarshal(pharmacyData, &data)
	if err != nil {
		println("Veri dönüşümünde sıkıntı var")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_DB_URL")))
	if err != nil {
		println("Hata Var")
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	for i := range data {
		dataFirst := data[i]
		many, err := client.Database(os.Getenv("MONGO_DB_NAME")).Collection(os.Getenv("MONGO_DB_COLLECTION_NAME")+strings.Split(time.Now().Format(time.RFC850), " ")[1], options.Collection()).InsertOne(context.Background(), dataFirst)
		println(data)
		if err != nil {
			return
		}
		println(many.InsertedID)
	}
}

func getAddressLatLng(address string) string {
	client, err := maps.NewClient(maps.WithAPIKey(os.Getenv("MAPS_API_KEY")))
	if err != nil {
		print("Error")
	}
	a, bias := client.FindPlaceFromText(context.Background(), &maps.FindPlaceFromTextRequest{
		Input:     address,
		InputType: maps.FindPlaceFromTextInputTypeTextQuery,
		Fields:    []maps.PlaceSearchFieldMask{"geometry"},
	})
	if bias != nil {
		print("Error")
	}
	result := a.Candidates
	if len(result) == 1 {
		return result[0].Geometry.Location.String()
	}
	return ""
}

func getFetchDataProvince(provinceName string) {

	turkishCharacterCapitalize := map[string]string{
		"adiyaman":      "Adıyaman",
		"agri":          "Ağrı",
		"aydin":         "Aydın",
		"balikesir":     "Balıkesir",
		"canakkale":     "Çanakkale",
		"cankiri":       "Çankırı",
		"corum":         "Çorum",
		"elazig":        "Elazığ",
		"eskisehir":     "Eskişehir",
		"gumushane":     "Gümüşhane",
		"isparta":       "Isparta",
		"istanbul":      "İstanbul",
		"izmir":         "İzmir",
		"kirklareli":    "Kırklareli",
		"kirsehir":      "Kırşehir",
		"kutahya":       "Kütahya",
		"kahramanmaras": "Kahramanmaraş",
		"mugla":         "Muğla",
		"mus":           "Muş",
		"nevsehir":      "Nevşehir",
		"nigde":         "Niğde",
		"tekirdag":      "Tekirdağ",
		"sanliurfa":     "Şanlıurfa",
		"usak":          "Uşak",
		"kirikkale":     "Kırıkkale",
		"sirnak":        "Şırnak",
		"bartin":        "Bartın",
		"igdir":         "Iğdır",
		"karabuk":       "Karabük",
		"duzce":         "Düzce",
	}

	pharmacyList := make([]Pharmacy, 0)
	c := colly.NewCollector()
	c.OnRequest(func(request *colly.Request) {
		request.ResponseCharacterEncoding = "windows-1254"
	})
	c.OnHTML("div#orta", func(element *colly.HTMLElement) {
		resultReplace := strings.ReplaceAll(element.ChildText("div.HhKk"), ",", "")
		eczaneReplace := strings.ReplaceAll(resultReplace, "Eczane-İlçe:", ",")
		adressReplace := strings.ReplaceAll(eczaneReplace, "Adres:", ",")
		karakterReplace := strings.ReplaceAll(adressReplace, "©", "")
		telefonReplace := strings.ReplaceAll(karakterReplace, "Telefon:", ",")

		result := strings.Split(telefonReplace, ",")

		district := ""
		pharmacyName := ""
		var pharmacyNumber = ""
		var pharmacyAddress = ""
		var pharmacyLatLng = ""
		var j = 1
		for i := 1; i < len(result); i++ {
			if i == j {

				// Pharmacy Name
				name := strings.Split(result[i], "  ")
				println("End")
				if provinceName == "izmir" || provinceName == "mugla" || provinceName == "osmaniye" {
					pharmacyName = name[1]
					district = name[2]
				} else {
					pharmacyName = name[0]
					district = name[1]
				}
				if provinceName == "erzurum" {
					pharmacyName = name[0]+" "+name[1]
					district = name[2]
				}

				j += 3
				println("Name : " + pharmacyName)
				println("District :" + district)
				// Pharmacy Number
				if len(result[i+2]) > 15 {
					if strings.Contains(result[i+2], "Saat") {
						pharmacyNumber = strings.Split(result[i+2], "Saat")[0]
					} else {
						pharmacyNumber = strings.Split(result[i+2], "Nöb")[0]
					}
				} else {
					pharmacyNumber = result[i+2]
				}

				// Pharmacy Address
				const nbsp = '\u00A0'
				if strings.Contains(result[i+1], string(nbsp)) {
					pharmacyAddress = strings.ReplaceAll(result[i+1], string(nbsp), "")
				} else {
					pharmacyAddress = result[i+1]
				}

				// latLng değerini tam olarak alabilmek için böyle yaptım.

				// Pharmacy LatLng
				if val, ok := turkishCharacterCapitalize[provinceName]; ok {
					splitPharmacyAddress := strings.SplitAfter(pharmacyAddress, val)
					pharmacyLatLng = getAddressLatLng(pharmacyName + " " + district + "/" + val)

					if pharmacyLatLng == "" {
						pharmacyLatLng = getAddressLatLng(strings.ReplaceAll(pharmacyAddress, "No", " No"))
					}
					if len(splitPharmacyAddress[0]) > len(provinceName) {
						pharmacyAddress = splitPharmacyAddress[0]
					}
				} else {
					splitPharmacyAddress := strings.SplitAfter(pharmacyAddress, strings.Title(provinceName))
					pharmacyLatLng = getAddressLatLng(pharmacyName + " " + district + "/" + strings.Title(provinceName))

					if pharmacyLatLng == "" {
						pharmacyLatLng = getAddressLatLng(strings.ReplaceAll(pharmacyAddress, "No", " No"))
					}

					if len(splitPharmacyAddress[0]) > len(provinceName) {
						pharmacyAddress = splitPharmacyAddress[0]
					}
				}
				pharmacy := Pharmacy{
					PharmacyName:        pharmacyName,
					PharmacyDistrict:    turkishCharacterProblem(district),
					PharmacyProvince:    provinceName,
					PharmacyAddress:     pharmacyAddress,
					PharmacyPhoneNumber: pharmacyNumber,
					PharmacyLatLng:      pharmacyLatLng,
				}
				pharmacyList = append(pharmacyList, pharmacy)
			}
		}
	})

	err := c.Visit(os.Getenv("VISIT_URL") + provinceName + os.Getenv("VISIT_URL_LAST"))
	if err != nil {
		print(err.Error())
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", " ")
	errorEncode := enc.Encode(pharmacyList)
	if errorEncode != nil {
		println("Encode da hata var")
	}

	data, errorData := json.MarshalIndent(pharmacyList, "", "")
	if errorData != nil {
		print("Hata Var")
	}
	connectAndWriteMongoDb(data)
	//_ = ioutil.WriteFile(provinceName+".json", data, 0666)
}

func turkishCharacterProblem(turkishWord string) string {
	turkishCharacterList := []string{"ğ", "Ğ", "ı", "İ", "ö", "Ö", "ü", "Ü", "ş", "Ş", "ç", "Ç"}
	englishCharacterList := []string{"g", "G", "i", "I", "o", "O", "u", "U", "s", "S", "c", "C"}
	//println(strings.Split(time.Now().Format(time.RFC850), " ")[1])

	str := turkishWord

	for i := 0; i < len(str)-1; i++ {
		for k := 0; k < len(englishCharacterList); k++ {
			if string([]rune(str)[i]) == turkishCharacterList[k] {
				str = strings.ReplaceAll(str, turkishCharacterList[k], englishCharacterList[k])
			}
		}
	}
	return str
}

//for i :=0; i< len(province) ; i++ {
//	jsonFile, err := os.Open(province[i]+".json")
//	if err != nil {
//		println("Json verisi yok")
//	}
//	byteValue , _ := ioutil.ReadAll(jsonFile)
//
//	connectAndWriteMongoDb(byteValue)
//}