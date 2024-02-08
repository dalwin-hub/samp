package models

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

var emptyObjectId primitive.ObjectID

type Vendor struct {
	Id                 primitive.ObjectID `bson:"_id,omitempty"`
	TenantId           primitive.ObjectID `bson:"tenantId,omitempty"`
	BusinessId         primitive.ObjectID `bson:"businessId,omitempty"`
	BusinessUnitId     primitive.ObjectID `bson:"businessUnitId,omitempty"`
	SchemaVersion      int32              `bson:"schemaVersion"`
	VendorCompanyName  string             `bson:"vendorCompanyName"`
	VendorTechnologies []string           `bson:"vendorTechnologies"`
	VendorDocuments    []Document         `bson:"vendorDocuments"`
	VendorContact      []Contact          `bson:"vendorContact"`
	VendorComments     string             `bson:"vendorComments"`
	CreatedBy          int                `bson:"createdBy,omitempty"`
	CreatedAt          time.Time          `bson:"createdAt,omitempty"`
	UpdatedBy          int                `bson:"updatedBy,omitempty"`
	UpdatedAt          time.Time          `bson:"updatedAt,omitempty"`
	IsDeleted          bool               `bson:"isDeleted"`
}

type Document struct {
	VendorDocumentsUniqueName string `bson:"vendorDocumentsUniqueName"`
	VendorDocumentsName       string `bson:"vendorDocumentsName"`
	VendorDocumentsUploadName string `bson:"vendorDocumentsUploadName"`
	VendorDocumentsLocation   string `bson:"vendorDocumentsLocation"`
	VendorDocumentsStatus     bool   `bson:"vendorDocumentsStatus"`
	VendorDocumentsIsDeleted  bool   `bson:"vendorDocumentsIsDeleted"`
}

type Contact struct {
	VendorContactId        int    `bson:"vendorContactId"`
	FirstName              string `bson:"firstName"`
	LastName               string `bson:"lastName"`
	ContactCountryCodeId   int32  `bson:"contactCountryCodeId"`
	ContactNumber          string `bson:"contactNumber"`
	Email                  string `bson:"email"`
	VendorContactIsDeleted bool   `bson:"vendorContactIsDeleted"`
}

type ResponseContact struct {
	VendorContactId int    `bson:"vendorContactId"`
	FirstName       string `bson:"firstName"`
	LastName        string `bson:"lastName"`
}

type VendorFilter struct {
	AvailableDetails   AvailableDetails `bson:"availableDetails"`
	VendorTechnologies []string         `bson:"vendorTechnologies"`
	FromDate           string           `bson:"fromDate"`
	ToDate             string           `bson:"toDate"`
}

// SortOptions        SortOptions      `bson:"sortOptions,omitempty"`

type AvailableDetails struct {
	VendorCompany bool `bson:"vendorCompany"`
	Technologies  bool `bson:"technologies"`
	ContactPerson bool `bson:"contactPerson"`
	ContactNumber bool `bson:"contactNumber"`
	Email         bool `bson:"email"`
}

type VendorContactUpdate struct {
	TenantId        primitive.ObjectID `bson:"tenantId"`
	BusinessId      primitive.ObjectID `bson:"businessId"`
	BusinessUnitId  primitive.ObjectID `bson:"businessUnitId"`
	SchemaVersion   int32              `bson:"schemaVersion"`
	VendorCompanyID primitive.ObjectID `bson:"vendorCompanyId"`
	VendorContact   []Contact          `bson:"vendorContact"`
	UpdatedBy       int                `bson:"updatedBy"`
	UpdatedAt       time.Time          `bson:"updatedAt"`
}

// type SortOptions struct {
// 	SortVariable string `bson:"sortVariable"`
// 	SortMethodId int32  `bson:"sortMethodId"`
// }

type Search struct {
	SearchValue string `bson:"searchValue"`
	// SortOptions SortOptions `bson:"sortOptions,omitempty"`
}

type Country struct {
	CountryId int `bson:"countryId"`
	StateId   int `bson:"stateId"`
}

type Mail struct {
	From    From      `json:"from"`
	To      []To      `json:"to"`
	Cc      []Cc      `json:"cc"`
	Bcc     []Bcc     `json:"bcc"`
	Subject string    `json:"subject"`
	Content []Content `json:"content"`
	ReplyTo []ReplyTo `json:"replyTo"`
}

type From struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type To struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type Cc struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type Bcc struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type ReplyTo struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type Content struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type Attachments struct {
	AttachmentFileName    string `json:"attachmentFileName"`
	AttachmentType        string `json:"attachmentType"`
	AttachmentDisposition string `json:"attachmentDisposition"`
}

type CreatedResponse struct {
	VendorId *mongo.InsertOneResult `json:"vendorId"`
}

// var homeViewColumns = bson.M{
// 	"vendorCompanyName":  1,
// 	"vendorTechnologies": 1,
// 	"createdBy":          1,
// 	"createdAt":          1,
// 	"vendorContacts": bson.M{
// 		"$arrayElemAt": []interface{}{"$vendorContact", -1},
// 		"$elemMatch":   []interface{}{"$vendorContactIsDeleted", false},
// 	},
// }

var homeViewColumns = bson.D{
	{"vendorCompanyName", 1},
	{"vendorTechnologies", 1},
	{"createdBy", 1},
	{"createdAt", 1},
	{"vendorContacts", bson.D{
		{"$arrayElemAt", []interface{}{
			bson.D{
				{"$filter", bson.D{
					{"input", "$vendorContact"},
					{"as", "contact"},
					{"cond", bson.D{
						{"$eq", []interface{}{"$$contact.vendorContactIsDeleted", false}},
					}},
				}},
			},
			-1,
		}},
	}},
}

// var quickViewColumns = bson.M{
// 	"vendorCompanyName":                         1,
// 	"vendorTechnologies":                        1,
// 	"createdBy":                                 1,
// 	"createdAt":                                 1,
// 	"updatedBy":                                 1,
// 	"updatedAt":                                 1,
// 	"vendorComments":                            1,
// 	"vendorContact.vendorContactId":             1,
// 	"vendorContact.firstName":                   1,
// 	"vendorContact.lastName":                    1,
// 	"vendorContact.contactCountryCodeId":        1,
// 	"vendorContact.contactNumber":               1,
// 	"vendorContact.email":                       1,
// 	"vendorDocuments.vendorDocumentsName":       1,
// 	"vendorDocuments.vendorDocumentsUploadName": 1,
// 	"vendorDocuments.vendorDocumentsLocation":   1,
// }

// var quickViewColumns = bson.M{
// 	"vendorCompanyName":                   1,
// 	"vendorTechnologies":                  1,
// 	"createdBy":                           1,
// 	"createdAt":                           1,
// 	"updatedBy":                           1,
// 	"updatedAt":                           1,
// 	"vendorComments":                      1,
// 	"vendorContact":                       1,
// 	"vendorDocuments.vendorDocumentsName": 1,
// 	"vendorDocuments.vendorDocumentsUploadName": 1,
// 	"vendorDocuments.vendorDocumentsLocation":   1,
// }

var quickViewColumns = bson.D{
	{"vendorCompanyName", 1},
	{"vendorTechnologies", 1},
	{"createdBy", 1},
	{"createdAt", 1},
	{"updatedBy", 1},
	{"updatedAt", 1},
	{"vendorComments", 1},
	{"vendorDocuments.vendorDocumentsName", 1},
	{"vendorDocuments.vendorDocumentsUploadName", 1},
	{"vendorDocuments.vendorDocumentsLocation", 1},
	{"vendorContact", bson.D{
		{"$filter", bson.D{
			{"input", "$vendorContact"},
			{"as", "contact"},
			{"cond", bson.D{
				{"$eq", []interface{}{"$$contact.vendorContactIsDeleted", false}},
			}},
		}},
	}},
}

type CountryMaster struct {
	MasterName  string        `bson:"masterName"`
	MasterFeeds []MasterFeeds `bson:"masterFeeds"`
}

type MasterFeeds struct {
	CountryId  int32  `bson:"countryId"`
	Name       string `bson:"name"`
	Phone_code string `bson:"phone_code"`
}

type Users struct {
	USERID    int    `json:"user_id"`
	USER_NAME string `json:"user_name"`
	EMAIL     string `json:"email"`
}

type VendorArray struct {
	VendorId []string `bson:"vendorId"`
}

var availableDetailsFilter = []string{"Vendor company", "Technologies", "Contact Person", "Contact Number", "Email"}

func (v *Vendor) Prepare() {
	v.VendorCompanyName = html.EscapeString(strings.TrimSpace(v.VendorCompanyName))
	for _, val := range v.VendorDocuments {
		val.VendorDocumentsUploadName = html.EscapeString(strings.TrimSpace(val.VendorDocumentsUploadName))
		val.VendorDocumentsLocation = html.EscapeString(strings.TrimSpace(val.VendorDocumentsLocation))
		val.VendorDocumentsName = html.EscapeString(strings.TrimSpace(val.VendorDocumentsName))
	}
	for _, val := range v.VendorContact {
		val.FirstName = html.EscapeString(strings.TrimSpace(val.FirstName))
		val.LastName = html.EscapeString(strings.TrimSpace(val.LastName))
		val.ContactNumber = html.EscapeString(strings.TrimSpace(val.ContactNumber))
		val.Email = html.EscapeString(strings.TrimSpace(val.Email))
	}
}
func (v *VendorContactUpdate) ValidateContact(action string) error {
	switch strings.ToLower(action) {
	case "create":
		if v.TenantId == emptyObjectId {
			log.Println("In vendors.go line 289,required Tenant Id")
			return errors.New("required Tenant Id")
		}
		if v.BusinessId == emptyObjectId {
			log.Println("In vendors.go line 293,required Business Id")
			return errors.New("required Business Id")
		}
		if v.BusinessUnitId == emptyObjectId {
			log.Println("In vendors.go line 297,required Business Unit Id")
			return errors.New("required Business Unit Id")
		}
		if v.SchemaVersion == 0 {
			log.Println("In vendors.go line 301,required Schema Version")
			return errors.New("required Schema Version")
		}
		if v.VendorCompanyID == emptyObjectId {
			log.Println("In vendors.go line 305,required Vendor Company Name")
			return errors.New("required Vendor Company Name")
		}
		if v.VendorContact == nil || len(v.VendorContact) == 0 {
			log.Println("In vendors.go line 309,required atleast one vendor contact ID")
			return errors.New("required atleast one Vendor Contact ID")
		}
		for _, val := range v.VendorContact {
			if val.Email == "" {
				log.Println("In vendors.go line 314,required contact email")
				return errors.New("required Contact Email")
			}
		}
		if v.UpdatedBy == 0 {
			log.Println("In vendors.go line 319,required updated user ID")
			return errors.New("required updated User ID")
		}
		return nil

	}
	return nil
}

// Validate does some mandatory checks on the master model
func (v *Vendor) Validate(action string) error {
	switch strings.ToLower(action) {
	case "create":
		v.CreatedAt = time.Now()
		v.UpdatedAt = time.Now()
		if v.TenantId == emptyObjectId {
			log.Println("In vendors.go line 335,required Tenant Id")
			return errors.New("required Tenant Id")
		}
		if v.BusinessId == emptyObjectId {
			log.Println("In vendors.go line 339,required Business Id")
			return errors.New("required Business Id")
		}
		if v.BusinessUnitId == emptyObjectId {
			log.Println("In vendors.go line 343,required Business unit Id")
			return errors.New("required Business Unit Id")
		}
		if v.SchemaVersion == 0 {
			log.Println("In vendors.go line 347,required schema version")
			return errors.New("required Schema Version")
		}
		if v.VendorCompanyName == "" {
			log.Println("In vendors.go line 351,required vendor company name")
			return errors.New("required Vendor Company Name")
		}
		if v.CreatedBy == 0 {
			log.Println("In vendors.go line 355,required created user's ID")
			return errors.New("required created user's ID")
		}
		if v.VendorContact == nil || len(v.VendorContact) == 0 {
			log.Println("In vendors.go line 359,required atleast one vendor contact ID")
			return errors.New("required atleast one Vendor Contact ID")
		}
		for _, val := range v.VendorContact {
			if val.Email == "" {
				log.Println("In vendors.go line 364,required Contact Email")
				return errors.New("required Contact Email")
			}
		}
		if v.VendorTechnologies == nil {
			log.Println("In vendors.go line 369,NULL is not accepted for Vendor Technologies")
			return errors.New("NULL is not accepted for Vendor Technologies")
		}
		if v.VendorDocuments == nil {
			v.VendorDocuments = []Document{}
		}
		return nil

	case "update":
		v.UpdatedAt = time.Now()
		if v.SchemaVersion == 0 {
			log.Println("In vendors.go line 380,required schema version")
			return errors.New("required Schema Version")
		}
		if v.VendorCompanyName == "" {
			log.Println("In vendors.go line 384,required vendor company name")
			return errors.New("required Vendor Company Name")
		}
		if v.UpdatedBy == 0 {
			log.Println("In vendors.go line 388,required created user's ID")
			return errors.New("required updated user's ID")
		}
		for _, val := range v.VendorContact {
			if val.Email == "" {
				log.Println("In vendors.go line 393,required contact email")
				return errors.New("required Contact Email")
			}
		}
		return nil

	default:
		if v.SchemaVersion == 0 {
			log.Println("In vendors.go line 401,required schema version")
			return errors.New("required Schema Version")
		}
		return nil
	}
	//return nil
}

// SaveVendor stores vendor data into db
func (v *Vendor) SaveVendor(db *mongo.Collection) (*Vendor, *mongo.InsertOneResult, error) {
	insertId, err := db.InsertOne(context.TODO(), v)
	if err != nil {
		log.Printf("In vendors.go line 413,Error while inserting vendor data: %v", err)
		return &Vendor{}, insertId, err
	}
	return v, insertId, nil
}

// FindAllVendors list all the master data
func (v *Vendor) FindAllVendors(db *mongo.Collection, tId primitive.ObjectID, bId primitive.ObjectID, buId primitive.ObjectID, sortOpt string, countryDb *mongo.Collection, mysqlDb *gorm.DB, limit int64, offset int64) ([]primitive.M, error) {

	var results []primitive.M

	query := bson.M{}
	query["tenantId"] = tId
	query["businessId"] = bId
	query["businessUnitId"] = buId
	query["isDeleted"] = false
	//query["vendorContacts.vendorContactIsDeleted"] = false

	var opts *options.FindOptions
	if sortOpt == "old" {
		opts = options.Find().SetProjection(homeViewColumns).SetSort(bson.M{"updatedAt": 1}).SetSkip(int64(offset)).SetLimit(int64(limit))
	} else {
		opts = options.Find().SetProjection(homeViewColumns).SetSort(bson.M{"updatedAt": -1}).SetSkip(int64(offset)).SetLimit(int64(limit))
	}

	// Set up the options for Find() method
	// opts = options.Find().SetProjection(homeViewColumns).SetSort(bson.D{{"createdAt", -1}})
	// if sortOpt.SortVariable != "" {
	// 	opts = options.Find().SetProjection(homeViewColumns).SetSort(bson.D{{sortOpt.SortVariable, sortOpt.SortMethodId}})
	// }

	cur, err := db.Find(context.TODO(), query, opts) //returns a *mongo.Cursor
	if err != nil {
		log.Printf("In vendors.go line 446,Error occured while  executing Find query: %v", err)
		return results, err

	}
	for cur.Next(context.TODO()) { //Next() gets the next document for corresponding cursor

		var elem primitive.M
		err := cur.Decode(&elem)
		if err != nil {
			log.Printf("In vendors.go line 455,Error occured while decoding document: %v", err)
			return results, err
		}

		contactCountryCodeId := elem["vendorContacts"].(primitive.M)["contactCountryCodeId"].(int32)
		var countryCodeValue string
		if contactCountryCodeId != 0 {
			countryCodeValue, err = GetCountryCodeValue(contactCountryCodeId, countryDb)
			if err != nil {
				log.Printf("In vendors.go line 464,Error occured while getting country code value: %v", err)
				return results, err
			}
		}

		date := elem["createdAt"].(primitive.DateTime)
		formatDate, err := formatDate(date)
		elem["createdAt"] = formatDate

		elem["vendorContacts"].(primitive.M)["contactCountryCodeId"] = countryCodeValue

		createdUserId := elem["createdBy"].(int32)
		elem["createdBy"], _, err = GetUserDetails(mysqlDb, int(createdUserId))
		if err != nil {
			log.Println("In vendors.go line 478,Error occurred in createdBy")
			return results, err
		}
		results = append(results, elem)
	}
	return results, err
}

// FindVendorByID query specific master with their ID
func (v *Vendor) FindVendorByID(db *mongo.Collection, id primitive.ObjectID, tId primitive.ObjectID, bId primitive.ObjectID, buId primitive.ObjectID, countryDb *mongo.Collection, mysqlDb *gorm.DB) (primitive.M, error) {

	var result primitive.M //  an unordered representation of a BSON document which is a Map

	query := bson.M{}
	query["_id"] = id
	query["tenantId"] = tId
	query["businessId"] = bId
	query["businessUnitId"] = buId
	query["isDeleted"] = false
	err := db.FindOne(context.TODO(), query, options.FindOne().SetProjection(quickViewColumns)).Decode(&result)
	if err != nil {
		log.Println("In vendors.go line 499,Error occurred in fetching findone")
		return result, err
	}

	//fmt.Println(result)
	for x := range result["vendorContact"].(primitive.A) {
		contactCountryCodeId, ok := result["vendorContact"].(primitive.A)[x].(primitive.M)["contactCountryCodeId"].(int32)
		if !ok {
			log.Println("In vendors.go line 507,Error occurred in fetching the contact country code Id")
			return result, err
		}

		var countryCodeValue string

		if contactCountryCodeId != 0 {
			countryCodeValue, err = GetCountryCodeValue(contactCountryCodeId, countryDb)
			if err != nil {
				log.Println("In vendors.go line 516,Error occurred in fetching the country code value")
				return result, err
			}
		}

		// update the value of "contactCountryCodeId"
		result["vendorContact"].(primitive.A)[x].(primitive.M)["contactCountryCodeId"] = countryCodeValue
	}

	createdDate := result["createdAt"].(primitive.DateTime)
	updatedDate := result["updatedAt"].(primitive.DateTime)

	createdFormatDate, err := formatDate(createdDate)
	result["createdAt"] = createdFormatDate
	if err != nil {
		log.Println("In vendors.go line 531,Error occurred on created format date")
		return result, err
	}

	updatedFormatDate, err := formatDate(updatedDate)
	result["updatedAt"] = updatedFormatDate
	if err != nil {
		log.Println("In vendors.go line 538,Error occurred on updated format date")
		return result, err
	}

	createdUserId := result["createdBy"].(int32)
	result["createdBy"], _, err = GetUserDetails(mysqlDb, int(createdUserId))
	if err != nil {
		log.Println("In vendors.go line 545,Error occurred on created user Id")
		return result, err
	}

	updatedUserId := result["updatedBy"].(int32)
	result["updatedBy"], _, err = GetUserDetails(mysqlDb, int(updatedUserId))
	if err != nil {
		log.Println("In vendors.go line 552,Error occurred on updated user Id")
		return result, err
	}

	return result, err
}

// FindVendorByID query specific master with their ID
func FindLastVendorContactByID(db *mongo.Collection, id primitive.ObjectID, tId primitive.ObjectID, bId primitive.ObjectID, buId primitive.ObjectID) (int, error) {

	var result primitive.M

	query := bson.M{}
	query["_id"] = id
	query["tenantId"] = tId
	query["businessId"] = bId
	query["businessUnitId"] = buId
	query["isDeleted"] = false

	err := db.FindOne(context.TODO(), query, options.FindOne().SetProjection(homeViewColumns)).Decode(&result)
	if err != nil {
		log.Printf("In vendors.go line 573,Error while  executing FindOne query: %v", err)
		return 0, err
	}

	vendorContacts := result["vendorContacts"].(primitive.M)
	vendorContactID := vendorContacts["vendorContactId"].(int32)
	return int(vendorContactID), err
}

// FindVendorByName query specific master with their ID
func (v *Vendor) FindVendorByName(db *mongo.Collection, name string, tId primitive.ObjectID, bId primitive.ObjectID, buId primitive.ObjectID) (int, error) {

	var result primitive.M

	regexPattern := fmt.Sprintf("(?i)^%s$", regexp.QuoteMeta(name))

	query := bson.M{}
	query["vendorCompanyName"] = bson.M{"$regex": regexPattern}
	query["tenantId"] = tId
	query["businessId"] = bId
	query["businessUnitId"] = buId
	query["isDeleted"] = false
	err := db.FindOne(context.TODO(), query).Decode(&result)
	if err != nil {
		log.Println("In vendors.go line 597,Error occurred on find vendor by name")
		return len(result), nil
	}
	return len(result), err
}

// DeleteVendor updates Vendor's delete log to true in db
func (v *Vendor) DeleteVendor(db *mongo.Collection, id primitive.ObjectID, tId primitive.ObjectID, bId primitive.ObjectID, buId primitive.ObjectID) (string, error) {

	in := bson.M{}
	in["isDeleted"] = true

	query := bson.M{}
	query["_id"] = id
	query["tenantId"] = tId
	query["businessId"] = bId
	query["businessUnitId"] = buId
	query["isDeleted"] = false

	out, err := db.UpdateOne(context.TODO(), query, bson.M{"$set": in})
	if err != nil {
		log.Println("In vendors.go line 618,Error occurred while deleting vendor")
		return "error", err
	}
	if out.MatchedCount == 0 {
		return "Vendor Not Found", errors.New("Vendor Not Found")
	}
	return "Vendor has been deleted", nil
}

// UpdateVendor updates Vendor details
func (v *Vendor) UpdateVendor(db *mongo.Collection, id primitive.ObjectID, tId primitive.ObjectID, bId primitive.ObjectID, buId primitive.ObjectID) (string, error) {

	//in := bson.M{}
	//var err error
	//in,err = bson.Marshal(v)
	query := bson.M{}
	query["_id"] = id
	query["tenantId"] = tId
	query["businessId"] = bId
	query["businessUnitId"] = buId
	query["isDeleted"] = false

	set := bson.M{}
	set["vendorCompanyName"] = v.VendorCompanyName
	set["vendorTechnologies"] = v.VendorTechnologies
	set["vendorComments"] = v.VendorComments
	set["vendorContact"] = v.VendorContact
	set["updatedBy"] = v.UpdatedBy
	set["updatedAt"] = time.Now()

	out, err := db.UpdateOne(context.TODO(), query, bson.M{"$set": set})
	if err != nil {
		log.Println("In vendors.go line 650,Error occurred while updating vendor")
		return "error", err
	}
	if out.MatchedCount == 0 {
		return "Vendor Not Found", errors.New("Vendor Not Found")
	}
	return "Vendor has been updated", nil
}

func (v *Vendor) UpdateHotlistVendor(db *mongo.Collection, id primitive.ObjectID, tId primitive.ObjectID, bId primitive.ObjectID, buId primitive.ObjectID) (string, error) {

	//in := bson.M{}
	//var err error
	//in,err = bson.Marshal(v)
	query := bson.M{}
	query["_id"] = id
	query["tenantId"] = tId
	query["businessId"] = bId
	query["businessUnitId"] = buId

	update := bson.M{}
	update["vendorCompanyName"] = v.VendorCompanyName
	update["vendorContact"] = v.VendorContact
	update["updatedBy"] = v.UpdatedBy
	update["updatedAt"] = time.Now()
	out, err := db.UpdateOne(context.TODO(), query, bson.M{"$set": update})
	if err != nil {
		log.Println("In vendors.go line 677,Error occurred while updating hotlist vendor")
		return "error", err
	}
	if out.MatchedCount == 0 {
		return "Vendor Not Found", errors.New("Vendor Not Found")
	}
	return "Vendor has been updated", nil
}

func (v *Vendor) GetTechnologies(db *mongo.Collection, tenantId primitive.ObjectID, businessId primitive.ObjectID, businessUnitId primitive.ObjectID) []string {
	var allArrays []primitive.M
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"tenantId":       tenantId,
				"businessId":     businessId,
				"businessUnitId": businessUnitId,
				"isDeleted":      false,
			},
		},
		{
			"$unwind": "$vendorTechnologies",
		},
		{
			"$group": bson.M{
				"_id": "$vendorTechnologies",
			},
		},
	}
	// Execute the aggregate query
	cursor, err := db.Aggregate(context.TODO(), pipeline)
	if err != nil {
		log.Panic(err)
	}
	for cursor.Next(context.TODO()) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			log.Panic(err)
		}
		allArrays = append(allArrays, result)
	}
	result := make([]string, len(allArrays))
	for i, v := range allArrays {
		result[i] = v["_id"].(string)
	}
	//fmt.Println(result)

	return result
}

func (v *VendorFilter) GetAvailableDetails(db *mongo.Collection) []string {
	return availableDetailsFilter
}

func (v *VendorFilter) FilterVendor(db *mongo.Collection, tenantId primitive.ObjectID, businessId primitive.ObjectID, businessUnitId primitive.ObjectID, countryDb *mongo.Collection, mysqlDb *gorm.DB) ([]primitive.M, error) {
	var result []primitive.M //  an unordered representation of a BSON document which is a Map
	filter := bson.M{}
	// oneWeekAgo := time.Now().Add(-7 * 24 * time.Hour).UTC()
	// twoWeekAgo := time.Now().Add(-14 * 24 * time.Hour).UTC()
	// sixMonthsAgo := time.Now().Add(-180 * 24 * time.Hour).UTC()
	filter["tenantId"] = tenantId
	filter["businessId"] = businessId
	filter["businessUnitId"] = businessUnitId
	filter["isDeleted"] = false

	if v.AvailableDetails.VendorCompany != false {
		filter["vendorCompanyName"] = bson.M{"$ne": nil}
		filter["vendorCompanyName"] = bson.M{"$ne": ""}
	}
	if v.AvailableDetails.Technologies != false {
		filter["vendorTechnologies"] = bson.M{"$ne": nil}
		filter["vendorTechnologies"] = bson.M{"$ne": ""}
	}
	if v.AvailableDetails.ContactPerson != false {
		filter["vendorContact.$.contactFirstName"] = bson.M{"$ne": nil}
		filter["vendorContact.$.contactFirstName"] = bson.M{"$ne": ""}

		filter["vendorContact.$.contactLastName"] = bson.M{"$ne": nil}
		filter["vendorContact.$.contactLastName"] = bson.M{"$ne": ""}
	}
	if v.AvailableDetails.ContactNumber != false {
		filter["vendorContact.$.contactCountryCodeId"] = bson.M{"$ne": nil}
		filter["vendorContact.$.contactCountryCodeId"] = bson.M{"$ne": ""}

		filter["vendorContact.$.contactNumber"] = bson.M{"$ne": nil}
		filter["vendorContact.$.contactNumber"] = bson.M{"$ne": ""}
	}
	if v.AvailableDetails.Email != false {
		filter["vendorContact.$.email"] = bson.M{"$ne": nil}
		filter["vendorContact.$.email"] = bson.M{"$ne": ""}
	}
	if len(v.VendorTechnologies) > 0 {
		filter["vendorTechnologies"] = bson.M{"$in": v.VendorTechnologies}
	}

	// Set up the options for Find() method
	opts := options.Find().SetProjection(homeViewColumns).SetSort(bson.D{{"updatedAt", 1}})
	// if v.SortOptions.SortVariable != "" {
	// 	opts = options.Find().SetProjection(homeViewColumns).SetSort(bson.D{{v.SortOptions.SortVariable, v.SortOptions.SortMethodId}})
	// }

	cur, err := db.Find(context.TODO(), filter, opts)
	if err != nil {
		return result, err
	}
	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
		var elem bson.M
		err := cur.Decode(&elem)
		if err != nil {
			return result, err
		}
		contactCountryCodeId := elem["vendorContacts"].(primitive.M)["contactCountryCodeId"].(int32)
		var countryCodeValue string
		if contactCountryCodeId != 0 {
			countryCodeValue, err = GetCountryCodeValue(contactCountryCodeId, countryDb)
		}
		if err != nil {
			return result, err
		}
		date := elem["createdAt"].(primitive.DateTime)
		// fmt.Println("date", date)
		formatDate, err := formatDate(date)
		elem["createdAt"] = formatDate
		// update the value of "contactCountryCodeId"
		elem["vendorContacts"].(primitive.M)["contactCountryCodeId"] = countryCodeValue

		createdUserId := elem["createdBy"].(int32)
		elem["createdBy"], _, err = GetUserDetails(mysqlDb, int(createdUserId))
		if err != nil {
			log.Println("In vendors.go line 808,Error occurred on get user details")
			return result, err
		}
		result = append(result, elem)
	}
	return result, nil
}

func (v *VendorFilter) FilterVendorTest(
	db *mongo.Collection,
	tenantId primitive.ObjectID,
	businessId primitive.ObjectID,
	businessUnitId primitive.ObjectID,
	countryDb *mongo.Collection,
	mysqlDb *gorm.DB,
	fromDate time.Time,
	toDate time.Time,
) ([]primitive.M, error) {
	var result []primitive.M // an unordered representation of a BSON document which is a Map
	filter := bson.M{}
	filter["tenantId"] = tenantId
	filter["businessId"] = businessId
	filter["businessUnitId"] = businessUnitId
	filter["isDeleted"] = false

	if v.AvailableDetails.VendorCompany != false {
		filter["vendorCompanyName"] = bson.M{"$ne": nil}
		filter["vendorCompanyName"] = bson.M{"$ne": ""}
	}
	if v.AvailableDetails.Technologies != false {
		filter["vendorTechnologies"] = bson.M{"$ne": nil}
		filter["vendorTechnologies"] = bson.M{"$ne": ""}
	}
	if v.AvailableDetails.ContactPerson != false {
		filter["vendorContact.$.contactFirstName"] = bson.M{"$ne": nil}
		filter["vendorContact.$.contactFirstName"] = bson.M{"$ne": ""}

		filter["vendorContact.$.contactLastName"] = bson.M{"$ne": nil}
		filter["vendorContact.$.contactLastName"] = bson.M{"$ne": ""}
	}
	if v.AvailableDetails.ContactNumber != false {
		filter["vendorContact.$.contactCountryCodeId"] = bson.M{"$ne": nil}
		filter["vendorContact.$.contactCountryCodeId"] = bson.M{"$ne": ""}

		filter["vendorContact.$.contactNumber"] = bson.M{"$ne": nil}
		filter["vendorContact.$.contactNumber"] = bson.M{"$ne": ""}
	}
	if v.AvailableDetails.Email != false {
		filter["vendorContact.$.email"] = bson.M{"$ne": nil}
		filter["vendorContact.$.email"] = bson.M{"$ne": ""}
	}
	if len(v.VendorTechnologies) > 0 {
		filter["vendorTechnologies"] = bson.M{"$in": v.VendorTechnologies}
	}

	//if !fromDate.IsZero() || !toDate.IsZero() {
	if !fromDate.IsZero() && !toDate.IsZero() {
		filter["createdAt"] = bson.M{
			"$gte": fromDate,
			"$lte": toDate,
		}

	} else if !fromDate.IsZero() {
		filter["createdAt"] = bson.M{
			"$gte": fromDate,
		}
	}

	// Set up the options for Find() method
	opts := options.Find().SetProjection(homeViewColumns).SetSort(bson.D{{"updatedAt", 1}})
	// if v.SortOptions.SortVariable != "" {
	// 	opts = options.Find().SetProjection(homeViewColumns).SetSort(bson.D{{v.SortOptions.SortVariable, v.SortOptions.SortMethodId}})
	// }

	cur, err := db.Find(context.TODO(), filter, opts)
	if err != nil {
		return result, err
	}
	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
		var elem bson.M
		err := cur.Decode(&elem)
		if err != nil {
			return result, err
		}
		contactCountryCodeId := elem["vendorContacts"].(primitive.M)["contactCountryCodeId"].(int32)
		var countryCodeValue string
		if contactCountryCodeId != 0 {
			countryCodeValue, err = GetCountryCodeValue(contactCountryCodeId, countryDb)
		}
		if err != nil {
			return result, err
		}
		date := elem["createdAt"].(primitive.DateTime)
		formatDate, err := formatDate(date)
		elem["createdAt"] = formatDate
		elem["vendorContacts"].(primitive.M)["contactCountryCodeId"] = countryCodeValue

		createdUserId := elem["createdBy"].(int32)
		elem["createdBy"], _, err = GetUserDetails(mysqlDb, int(createdUserId))
		if err != nil {
			log.Println("In vendors.go line 910,Error occurred on get user details")
			return result, err
		}
		result = append(result, elem)
	}
	return result, nil

}

func ParseDateParameters(fromDate, toDate string) (fromTime time.Time, toTime time.Time, err error) {
	// Parse fromDate
	if fromDate == "" {
		fromTime = time.Time{}
	} else {
		fromTime, err = time.Parse("2006-01-02", fromDate)
		if err != nil {
			return fromTime, toTime, err
		}
	}
	// Parse toDate
	if toDate == "" {
		toTime = time.Time{}
	} else {
		toTime, err = time.Parse("2006-01-02", toDate)
		if err != nil {
			return fromTime, toTime, err
		}
	}

	return fromTime, toTime, nil
}

// GetCountryList Fetches The Country List
func GetCountryList(db *mongo.Collection) ([]primitive.M, error) {
	pipeline := bson.A{
		bson.M{
			"$match": bson.M{"masterName": "Country"},
		},
		bson.M{
			"$unwind": "$masterFeeds",
		},
		bson.M{
			"$project": bson.M{
				"_id":       0,
				"countryId": "$masterFeeds.countryId",
				"name":      "$masterFeeds.name",
			},
		},
	}
	cursor, err := db.Aggregate(context.Background(), pipeline)
	if err != nil {
		log.Panic(err)
	}
	defer cursor.Close(context.Background())

	var results []bson.M
	if err = cursor.All(context.Background(), &results); err != nil {
		log.Panic(err)
	}
	return results, err

}

// GetStateList fetches states available inside a country
func GetStateList(db *mongo.Collection, countryId int) ([]primitive.M, error) {
	pipeline := bson.A{
		bson.M{
			"$unwind": "$masterFeeds",
		},
		bson.M{
			"$match": bson.M{
				"masterName":            "State",
				"masterFeeds.countryId": countryId,
			},
		},
		bson.M{
			"$project": bson.M{
				"_id":     0,
				"stateId": "$masterFeeds.stateId",
				"name":    "$masterFeeds.name",
			},
		},
	}
	cursor, err := db.Aggregate(context.Background(), pipeline)
	if err != nil {
		log.Panic(err)
	}
	defer cursor.Close(context.Background())

	var results []bson.M
	if err = cursor.All(context.Background(), &results); err != nil {
		log.Panic(err)
	}
	return results, err

}

// GetCityList fetches states available inside a state
func GetCityList(db *mongo.Collection, countryId int, stateId int) ([]primitive.M, error) {
	pipeline := bson.A{
		bson.M{
			"$unwind": "$masterFeeds",
		},
		bson.M{
			"$match": bson.M{
				"masterName":          "City",
				"countryId":           countryId,
				"masterFeeds.stateId": stateId,
			},
		},
		bson.M{
			"$project": bson.M{
				"_id":    0,
				"cityId": "$masterFeeds.cityId",
				"name":   "$masterFeeds.name",
			},
		},
	}
	cursor, err := db.Aggregate(context.Background(), pipeline)
	if err != nil {
		log.Panic(err)
	}
	defer cursor.Close(context.Background())

	var results []bson.M
	if err = cursor.All(context.Background(), &results); err != nil {
		log.Panic(err)
	}
	return results, err

}

func GetCountryCode(db *mongo.Collection) ([]primitive.M, error) {

	pipeline := bson.A{
		bson.M{
			"$match": bson.M{"masterName": "Country"},
		},
		bson.M{
			"$unwind": "$masterFeeds",
		},
		bson.M{
			"$project": bson.M{
				"_id":       0,
				"countryId": "$masterFeeds.countryId",
				"phoneCode": "$masterFeeds.phone_code",
			},
		},
	}
	cursor, err := db.Aggregate(context.Background(), pipeline)
	if err != nil {
		log.Panic(err)
	}
	defer cursor.Close(context.Background())

	var results []bson.M
	if err = cursor.All(context.Background(), &results); err != nil {
		log.Panic(err)
	}
	return results, err
}

func GetCountryCodeValue(countryID int32, db *mongo.Collection) (string, error) {
	//fmt.Println("inside fun")
	filter := bson.M{"masterFeeds.countryId": countryID}
	projection := bson.M{"masterFeeds.$": 1}
	var result CountryMaster
	err := db.FindOne(context.Background(), filter, options.FindOne().SetProjection(projection)).Decode(&result)
	if err != nil {
		log.Println("In vendors.go line 1079,Error occurred on get country code value")
		return "", nil
	}

	// Extract the country name from the result
	var countryCode string
	for _, feed := range result.MasterFeeds {
		if feed.CountryId == countryID {
			countryCode = feed.Phone_code
			break
		}
	}
	//fmt.Println("countryName", countryName)

	return countryCode, nil

}

func formatDate(dateStr primitive.DateTime) (string, error) {
	// Parse the input date string into a time.Time object
	t := time.UnixMilli(int64(dateStr))
	formattedDate := t.Format("January 02, 2006")

	return formattedDate, nil
}

func (v *Vendor) EditGetVendor(db *mongo.Collection, id primitive.ObjectID, tId primitive.ObjectID, bId primitive.ObjectID, buId primitive.ObjectID) (primitive.M, error) {
	var result primitive.M //  an unordered representation of a BSON document which is a Map
	query := bson.M{}
	query["_id"] = id
	query["tenantId"] = tId
	query["businessId"] = bId
	query["businessUnitId"] = buId
	query["isDeleted"] = false
	err := db.FindOne(context.TODO(), query, options.FindOne().SetProjection(quickViewColumns)).Decode(&result)
	if err != nil {
		log.Println("In vendors.go line 1115,Error occurred on edit get vendor")
		return result, err
	}

	createdAt := result["createdAt"].(primitive.DateTime)
	// fmt.Println("date", date)
	formatcreatedAt, err := formatDate(createdAt)
	result["createdAt"] = formatcreatedAt
	// updatedAt := result["updatedAt"].(primitive.DateTime)
	// // fmt.Println("date", date)
	// formatupdatedAt, err := formatDate(updatedAt)
	// result["updatedAt"] = formatupdatedAt

	return result, err
}

func (m *Mail) VendorMailShare(tenantId primitive.ObjectID, businessId primitive.ObjectID, businessUnitId primitive.ObjectID) (string, error) {
	postBody, _ := json.Marshal(map[string]interface{}{
		"from":    m.From,
		"to":      m.To,
		"cc":      m.Cc,
		"subject": m.Subject,
		"content": m.Content,
		"replyTo": m.ReplyTo,
	})
	responseBody := bytes.NewBuffer(postBody)
	resp1, err := http.Post("http://192.168.1.22:7000/api/notify", "application/json", responseBody)
	//Handle Error
	if err != nil {
		log.Panicf("An Error Occured %v", err)
	}
	defer resp1.Body.Close()
	//Read the response body
	if resp1.StatusCode == 200 {
		return "Vendor Shared Successfully", nil
	} else {
		return "Vendor not shared", nil
	}
}

func (s *Search) VendorSearch(vendorDb *mongo.Collection, tenantId primitive.ObjectID, businessId primitive.ObjectID, businessUnitId primitive.ObjectID, sortOpt string, countryDb *mongo.Collection, mysqlDb *gorm.DB) ([]primitive.M, error) {
	var result []primitive.M //  an unordered representation of a BSON document which is a Map

	filter := bson.M{
		"$or": []bson.M{
			{"vendorCompanyName": bson.M{"$regex": s.SearchValue, "$options": "i"}},
			{"vendorTechnologies": bson.M{"$elemMatch": bson.M{"$regex": s.SearchValue, "$options": "i"}}},
			{"vendorComments": bson.M{"$regex": s.SearchValue, "$options": "i"}},
			{"vendorContact.firstName": bson.M{"$regex": s.SearchValue, "$options": "i"}},
			{"vendorContact.lastName": bson.M{"$regex": s.SearchValue, "$options": "i"}},
			{"vendorContact.contactNumber": bson.M{"$regex": s.SearchValue, "$options": "i"}},
			{"vendorContact.email": bson.M{"$regex": s.SearchValue, "$options": "i"}},
			{"createdAt": bson.M{"$regex": s.SearchValue, "$options": "i"}},
			{"updatedAt": bson.M{"$regex": s.SearchValue, "$options": "i"}},
		},
		"tenantId":       tenantId,
		"businessId":     businessId,
		"businessUnitId": businessUnitId,
		"isDeleted":      false,
	}

	// Set up the options for Find() method
	// opts := options.Find().SetProjection(homeViewColumns).SetSort(bson.D{{"createdAt", -1}})
	// if s.SortOptions.SortVariable != "" {
	// 	opts = options.Find().SetProjection(homeViewColumns).SetSort(bson.D{{s.SortOptions.SortVariable, s.SortOptions.SortMethodId}})
	// }
	var opts *options.FindOptions
	if sortOpt == "old" {
		opts = options.Find().SetProjection(homeViewColumns).SetSort(bson.M{"createdAt": 1})
	} else {
		opts = options.Find().SetProjection(homeViewColumns).SetSort(bson.M{"createdAt": -1})
	}
	cur, err := vendorDb.Find(context.TODO(), filter, opts)
	if err != nil {
		log.Println("In vendors.go line 1189,Error occurred on vendor serach")
		return result, err
	}
	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
		var elem bson.M
		err := cur.Decode(&elem)
		if err != nil {
			return result, err
		}
		contactCountryCodeId := elem["vendorContacts"].(primitive.M)["contactCountryCodeId"].(int32)
		var countryCodeValue string
		if contactCountryCodeId != 0 {
			countryCodeValue, err = GetCountryCodeValue(contactCountryCodeId, countryDb)
		}
		if err != nil {
			log.Println("In vendors.go line 1206,Error occurred on contact countrycode Id")
			return result, err
		}
		date := elem["createdAt"].(primitive.DateTime)
		// fmt.Println("date", date)
		formatDate, err := formatDate(date)
		elem["createdAt"] = formatDate
		// update the value of "contactCountryCodeId"
		elem["vendorContacts"].(primitive.M)["contactCountryCodeId"] = countryCodeValue

		createdUserId := elem["createdBy"].(int32)
		elem["createdBy"], _, err = GetUserDetails(mysqlDb, int(createdUserId))
		if err != nil {
			log.Println("In vendors.go line 1219,Error occurred on created user Id")
			return result, err
		}

		result = append(result, elem)
	}
	return result, nil
}

// FindAllVendors list all the master data
func FetchVendorNames(db *mongo.Collection, tId primitive.ObjectID, bId primitive.ObjectID, buId primitive.ObjectID) ([]map[string]interface{}, error) {

	query := bson.M{}
	query["tenantId"] = tId
	query["businessId"] = bId
	query["businessUnitId"] = buId
	query["isDeleted"] = false

	// Projection to include _id and vendorCompanyName fields
	projection := bson.M{
		"_id":               1,
		"vendorCompanyName": 1,
	}

	// Find the documents and retrieve vendor names with _id
	cursor, err := db.Find(context.Background(), query, options.Find().SetProjection(projection))
	if err != nil {
		log.Panic(err)
	}
	defer cursor.Close(context.Background())

	var vendors []Vendor
	if err := cursor.All(context.Background(), &vendors); err != nil {
		log.Panic(err)
	}

	// Extract vendor names with _id from the result
	var vendorData []map[string]interface{}
	for _, vendor := range vendors {
		vendorData = append(vendorData, map[string]interface{}{
			"_id":               vendor.Id.Hex(),
			"vendorCompanyName": vendor.VendorCompanyName,
		})
	}

	fmt.Println("Vendor Data:", vendorData)

	return vendorData, err
}

// update vendor contact
func UpdateContactInfoForVendor(db *mongo.Collection, contactDetails VendorContactUpdate) error {

	var response *mongo.UpdateResult
	var err error

	filter := bson.M{}
	filter["tenantId"] = contactDetails.TenantId
	filter["businessId"] = contactDetails.BusinessId
	filter["businessUnitId"] = contactDetails.BusinessUnitId
	filter["_id"] = contactDetails.VendorCompanyID
	filter["isDeleted"] = false

	push := bson.M{}
	push["vendorContact"] = contactDetails.VendorContact

	update := bson.M{}
	update["updatedBy"] = contactDetails.UpdatedBy
	update["updatedAt"] = time.Now()

	for x := range contactDetails.VendorContact {
		push["vendorContact"] = contactDetails.VendorContact[x]
		response, err = db.UpdateOne(context.TODO(), filter, bson.M{"$set": update, "$push": push})
		if err != nil {
			log.Println("In vendors.go line 1293,Error occurred on vendor contact")
			return err
		}
	}

	if response.MatchedCount == 0 {
		return errors.New("no record found")
	}

	return nil

}

// FetchVendorContactNames gets all vendor contact names for the provided vendor id
func FetchVendorContactNames(db *mongo.Collection, tId primitive.ObjectID, bId primitive.ObjectID, buId primitive.ObjectID, vendorId primitive.ObjectID) ([]map[string]interface{}, error) {
	var contacts []map[string]interface{}
	query := bson.M{}
	query["tenantId"] = tId
	query["businessId"] = bId
	query["businessUnitId"] = buId
	query["_id"] = vendorId
	query["isDeleted"] = false

	// Projection to include _id and vendorCompanyName fields
	projection := bson.M{
		"vendorContact.vendorContactId":      1,
		"vendorContact.firstName":            1,
		"vendorContact.lastName":             1,
		"vendorContact.contactCountryCodeId": 1,
		"vendorContact.contactNumber":        1,
		"vendorContact.email":                1,
	}

	// Find the document and retrieve the desired fields
	var vendor Vendor
	err := db.FindOne(context.Background(), query, options.FindOne().SetProjection(projection)).Decode(&vendor)
	if err != nil {
		log.Print(err)
		return contacts, err
	}

	// Extract the desired fields as a list of objects

	for _, contact := range vendor.VendorContact {
		contactData := map[string]interface{}{
			"vendorContactId":      contact.VendorContactId,
			"firstName":            contact.FirstName,
			"lastName":             contact.LastName,
			"contactCountryCodeId": contact.ContactCountryCodeId,
			"contactNumber":        contact.ContactNumber,
			"email":                contact.Email,
		}
		contacts = append(contacts, contactData)
	}

	//fmt.Println("Vendor Data:", contacts)

	return contacts, err
}

// FindVendorContactByEmail query specific master with their ID
func FindVendorContactByEmail(db *mongo.Collection, tId primitive.ObjectID, bId primitive.ObjectID, buId primitive.ObjectID, vendorId primitive.ObjectID, email string) (int, error) {
	var result primitive.M //  an unordered representation of a BSON document which is a Map
	query := bson.M{}
	query["tenantId"] = tId
	query["businessId"] = bId
	query["businessUnitId"] = buId
	query["_id"] = vendorId
	query["vendorContact.email"] = email
	query["isDeleted"] = false
	err := db.FindOne(context.TODO(), query).Decode(&result)
	if err != nil {
		log.Println("In vendors.go line 1365,Error occurred on find vendor contact by email")
		return len(result), nil
	}
	return len(result), err
}

// UpdateVendorDocuments updates Vendor Documents
func UpdateVendorDocuments(db *mongo.Collection, tId primitive.ObjectID, bId primitive.ObjectID, buId primitive.ObjectID, id primitive.ObjectID, document []Document) (string, error) {

	var response *mongo.UpdateResult
	var err error
	push := bson.M{}
	query := bson.M{}
	query["_id"] = id
	query["tenantId"] = tId
	query["businessId"] = bId
	query["businessUnitId"] = buId
	query["isDeleted"] = false

	for x := range document {
		push["vendorDocuments"] = document[x]
		response, err = db.UpdateOne(context.TODO(), query, bson.M{"$push": push})
		if err != nil {
			log.Println("In vendors.go line 1388,Error occurred on vendor documents")
			return "error", err
		}
	}

	if response.MatchedCount == 0 {
		return "Vendor Not Found", errors.New("Vendor Not Found")
	}
	return "Vendor has been updated", nil
}

// UpdateVendorDocuments updates Vendor Documents
func PushVendorDocuments(db *mongo.Collection, tId primitive.ObjectID, bId primitive.ObjectID, buId primitive.ObjectID, id primitive.ObjectID, document Vendor) (string, error) {

	var response *mongo.UpdateResult
	var err error

	query := bson.M{}
	query["_id"] = id
	query["tenantId"] = tId
	query["businessId"] = bId
	query["businessUnitId"] = buId
	query["isDeleted"] = false

	set := bson.M{}
	set["vendorDocuments"] = document.VendorDocuments

	response, err = db.UpdateOne(context.TODO(), query, bson.M{"$set": set})
	if err != nil {
		log.Println("In vendors.go line 1417,Error occurred while pushing vendor documents ")
		return "error", err
	}

	if response.MatchedCount == 0 {
		return "Vendor Not Found", errors.New("Vendor Not Found")
	}
	return "Vendor has been updated", nil
}

// provides the user name and email using the provided userId
func GetUserDetails(db *gorm.DB, userId int) (string, string, error) {
	var users Users
	result := db.Table("ZNNXT_USER_MST").Select("USER_NAME as user_name, EMAIL as email").Where("USER_ID = ?", userId).Scan(&users)
	if result.Error != nil {
		log.Println("In vendors.go line 1432,Error occurred while getting user details")
		return "", "", result.Error
	}
	return users.USER_NAME, users.EMAIL, nil
}
