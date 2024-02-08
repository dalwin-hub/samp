package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"rltk-be-vendor/db"
	"rltk-be-vendor/models"
	"rltk-be-vendor/s3Bucket"
	"rltk-be-vendor/utils"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	VendorColl  = "vendors"
	CountryColl = "countryMaster"
	masterColl  = "masters"
)

var emptyObjectId primitive.ObjectID

// CreateVendor controller handles the master create API
// and returns the master created
func CreateVendor(w http.ResponseWriter, r *http.Request) {
	log := utils.GetLogger()

	userDetails, err := GetUsers(w, r)
	if err != nil {
		log.Println("Error on getting user details in vendorcontroller line 39: ", err)
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	err = r.ParseMultipartForm(32 << 20) // parse up to 32 MB of data
	if err != nil {
		log.Println("Error while unmarshalling JSON in vendorcontroller line 46: ", err)
		utils.ERROR(w, http.StatusUnprocessableEntity, err)
	}
	jsonStr := r.FormValue("json")

	// body, err := ioutil.ReadAll(r.Body)
	// if err != nil {
	// 	log.Warn(err)
	// 	utils.ERROR(w, http.StatusUnprocessableEntity, err)
	// }

	vendor := models.Vendor{}
	err = json.Unmarshal([]byte(jsonStr), &vendor)
	if err != nil {
		log.Warn("In vendorcontroller line 60,Error in unmarshalling json:", err)
		utils.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	vendor.TenantId, err = primitive.ObjectIDFromHex(userDetails.TENANT_ID)
	if err != nil {
		log.Warn("In vendorcontroller line 67,Error while converting TENANT_ID to ObjectID:", err)
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	vendor.BusinessId, err = primitive.ObjectIDFromHex(userDetails.BUSINESS_ID)
	if err != nil {
		log.Warn("In vendorcontroller line 74,Error while converting BUSINESS_ID to ObjectID:", err)
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	vendor.BusinessUnitId, err = primitive.ObjectIDFromHex(userDetails.BUSINESS_UNIT_ID)
	if err != nil {
		log.Warn("In vendorcontroller line 81,Error while converting BUSINESSUNIT_ID to ObjectID:", err)
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	vendor.CreatedBy = userDetails.USER_ID
	vendor.UpdatedBy = userDetails.USER_ID

	vendor.Prepare()

	err = vendor.Validate("create")
	if err != nil {
		log.Println("In vendorcontroller line 93,Error while validating create:", err)
		utils.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	//Check to verify if the Vendor Exists by using the Vendor Company Name
	countRcvd, err := vendor.FindVendorByName(db.GetCollection(VendorColl), vendor.VendorCompanyName, vendor.TenantId, vendor.BusinessId, vendor.BusinessUnitId)
	if err != nil {
		log.Println("In vendorController line 101, Error occured while finding vendor by name: ", err)
		utils.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	var emptyId *mongo.InsertOneResult
	var insertId *mongo.InsertOneResult
	if countRcvd == 0 {
		count := 1
		for x := range vendor.VendorContact {
			vendor.VendorContact[x].VendorContactId = count
			count += 1
		}
		_, insertId, err = vendor.SaveVendor(db.GetCollection(VendorColl))
		if err != nil {
			log.Println("In vendorcontroller line 116,Error occured while saving vendor: ", err)
			utils.ERROR(w, http.StatusInternalServerError, err)
			return
		}
	} else {
		utils.FailureResp(w, http.StatusOK, "vendor already exists")
		return
	}

	var documentArray []models.Document
	if insertId != emptyId {
		//upload resume
		documents := r.MultipartForm.File["documents"]
		// resumeFiles := r.FormValue.File[]
		if len(documents) != 0 {
			vendorInsertIdString := ExtractObjectID(fmt.Sprintf("%v", insertId))
			CreateUniqueResumeName := CreateUniqueFileName(documents, vendorInsertIdString, emptyObjectId, "vendor_document")

			vendorDocument := models.Document{}
			//upload the resume in s3 bucket
			for i, file := range documents {
				s3Bucket.MultiDocsUpload([]*multipart.FileHeader{file}, "vendor/documents/", CreateUniqueResumeName[i])
				vendorDocument.VendorDocumentsLocation = fmt.Sprintf("vendor/documents/%s", CreateUniqueResumeName[i])
				vendorDocument.VendorDocumentsUploadName = documents[i].Filename
				vendorDocument.VendorDocumentsUniqueName = CreateUniqueResumeName[i]
				vendorDocument.VendorDocumentsStatus = true
				documentArray = append(documentArray, vendorDocument)
			}
			vendorOid := insertId.InsertedID.(primitive.ObjectID)
			_, err := models.UpdateVendorDocuments(db.GetCollection(VendorColl), vendor.TenantId, vendor.BusinessId, vendor.BusinessUnitId, vendorOid, documentArray)
			if err != nil {
				log.Println("In vendorcontroller line 147,Error occured while updating vendor documents: ", err)
				utils.ERROR(w, http.StatusInternalServerError, err)
				return
			}
		}
	}

	//fmt.Println(documentArray)

	utils.SuccessResp(w, http.StatusOK, insertId)

}

func CreateUniqueFileName(Files []*multipart.FileHeader, vendorInsertIdString string, applicantId primitive.ObjectID, docName string) []string {
	//fmt.Println("CreateUniqueFileName", CreateUniqueFileName)
	// fileExtension := filepath.Ext(Files[0].Filename)
	//get the upload time
	var uniqueNames []string
	currentTime := time.Now().Format("20060102150405")

	// Convert the insertOneResult to a string
	//vendorInsertIdString := ExtractObjectID(fmt.Sprintf("%v", vendorInsertId))
	for _, file := range Files {
		fileExtension := filepath.Ext(file.Filename)
		//fmt.Println("fileExtension", fileExtension)
		//create the unique name to upload the resume
		uniqueName := fmt.Sprintf("%s_%s_%s%s", docName, vendorInsertIdString, currentTime, fileExtension)
		uniqueNames = append(uniqueNames, uniqueName)
	}

	return uniqueNames
}

func ExtractObjectID(input string) string {
	startIndex := strings.Index(input, `("`) + 2
	endIndex := strings.LastIndex(input, `")`)
	if startIndex == -1 || endIndex == -1 || endIndex <= startIndex {
		return ""
	}
	return input[startIndex:endIndex]
}

// GetVendors controller handles Vendors get api
// and returns all the Vendors
func GetVendors(w http.ResponseWriter, r *http.Request) {
	log := utils.GetLogger()
	Vendor := models.Vendor{}
	parameters := r.URL.Query()

	userDetails, err := GetUsers(w, r)
	if err != nil {
		log.Println("In vendorcontroller line 198,Error while getting user details:", err)
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	tenantId, err := primitive.ObjectIDFromHex(userDetails.TENANT_ID)
	if err != nil {
		log.Println("In vendorcontroller line 205,Error while converting TENANT_ID to ObjectID:", err)
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	businessId, err := primitive.ObjectIDFromHex(userDetails.BUSINESS_ID)
	if err != nil {
		log.Println("In vendorcontroller line 212,Error while converting BUSINESS_ID to ObjectID:", err)
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	businessUnitId, err := primitive.ObjectIDFromHex(userDetails.BUSINESS_UNIT_ID)
	if err != nil {
		log.Println("In vendorcontroller line 219,Error while converting BUSINESS_UNIT_ID to ObjectID:", err)
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	limitParam := parameters.Get("limit")
	limit, err := strconv.ParseInt(limitParam, 10, 16)
	if err != nil {
		log.Warn("In vendorcontroller line 227,Error occured while parsing limit parameter:", err)
		utils.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	offsetParam := parameters.Get("offset")
	offset, err := strconv.ParseInt(offsetParam, 10, 16)
	if err != nil {
		log.Warn("In vendorcontroller line 235,Error occured while parsing offset parameter:", err)
		utils.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	// sort := models.SortOptions{}

	// body, err := ioutil.ReadAll(r.Body)
	// if err != nil {
	// 	utils.ERROR(w, http.StatusBadRequest, err)
	// }

	// if len(body) != 0 {
	// 	err = json.Unmarshal(body, &sort)
	// 	if err != nil {
	// 		utils.ERROR(w, http.StatusBadRequest, err)
	// 		return
	// 	}
	// }

	sort := parameters.Get("sortBy")
	results, err := Vendor.FindAllVendors(db.GetCollection(VendorColl), tenantId, businessId, businessUnitId, sort, db.GetCollection(CountryColl), db.GetDB(), limit, offset)
	if err != nil {
		log.Warn("In vendorcontroller line 258,Error occured while parsing sortBy:", err)
		utils.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	// if results != nil {
	// 	utils.JSON(w, http.StatusCreated, vendors)
	// } else {
	// 	utils.JSON(w, http.StatusCreated, "No Records Found")
	// }

	// results, err := result.FindAllVendors(db.GetDB(VendorColl), tenantId, businessId, businessUnitId, sort, db.GetDB(CountryColl))
	// if err != nil {
	// 	utils.ERROR(w, http.StatusInternalServerError, err)
	// 	return
	// }
	utils.SuccessResp(w, http.StatusOK, results)
}

// GetVendor controller handles the Vendor get api
// and returns the requested Vendor data
func GetVendor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 283,Error occured while converting ID from hex")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	userDetails, err := GetUsers(w, r)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 290,Error occured while getting user details")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	tenantId, err := primitive.ObjectIDFromHex(userDetails.TENANT_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 297,Error occured while converting TENANT_ID from hex")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	businessId, err := primitive.ObjectIDFromHex(userDetails.BUSINESS_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 304,Error occured while converting BUSINESS_ID from hex")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	businessUnitId, err := primitive.ObjectIDFromHex(userDetails.BUSINESS_UNIT_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 311,Error occured while converting BUSINESS_UNIT_ID from hex")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	master := models.Vendor{}

	masterGotten, err := master.FindVendorByID(db.GetCollection(VendorColl), id, tenantId, businessId, businessUnitId, db.GetCollection(CountryColl), db.GetDB())
	if err != nil {
		utils.ERROR(w, http.StatusInternalServerError, err)
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 321,Invalid vendor ID")
		return
	}
	utils.SuccessResp(w, http.StatusOK, masterGotten)
}

func EditGetVendor(w http.ResponseWriter, r *http.Request) {
	vendor := models.Vendor{}
	vars := mux.Vars(r)
	idStr := vars["id"]
	_id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 333,Error occured while converting ID from hex")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	userDetails, err := GetUsers(w, r)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 340,Error occured while getting user details")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	tenantId, err := primitive.ObjectIDFromHex(userDetails.TENANT_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 347,Error occured while converting TENANT_ID from hex")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	businessId, err := primitive.ObjectIDFromHex(userDetails.BUSINESS_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 354,Error occured while converting BUSINESS_ID from hex")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	businessUnitId, err := primitive.ObjectIDFromHex(userDetails.BUSINESS_UNIT_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 361,Error occured while converting BUSINESS_UNIT_ID from hex")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	hotlists, err := vendor.EditGetVendor(db.GetCollection(VendorColl), _id, tenantId, businessId, businessUnitId)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 368,Error occured while fetching EditGetVendor")
		utils.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if hotlists != nil {
		utils.SuccessResp(w, http.StatusOK, hotlists)
	} else {
		utils.FailureResp(w, http.StatusOK, "Data not found")
	}
}

// DeleteVendor controller handles Vendor delete api
// and returns success or failure status
func DeleteVendor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 386,Error converting ID from hex")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	userDetails, err := GetUsers(w, r)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 393,Error getting user details")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	tenantId, err := primitive.ObjectIDFromHex(userDetails.TENANT_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 400,Error converting TENANT_ID from hex")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	businessId, err := primitive.ObjectIDFromHex(userDetails.BUSINESS_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 407,Error converting BUSINESS_ID from hex")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	businessUnitId, err := primitive.ObjectIDFromHex(userDetails.BUSINESS_UNIT_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 414,Error converting BUSINESS_UNIT_ID from hex")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	vendor := models.Vendor{}
	//tokenID, err := app.ExtractTokenID(r)
	//if err != nil {
	//	utils.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
	//	return
	//}
	//if tokenID != 0 && tokenID != uint32(uid) {
	//	utils.ERROR(w, http.StatusUnauthorized, errors.New(http.StatusText(http.StatusUnauthorized)))
	//	return
	//}
	response, err := vendor.DeleteVendor(db.GetCollection(VendorColl), id, tenantId, businessId, businessUnitId)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 431,Error occured in DeleteVendor")
		utils.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if response == "Vendor Not Found" {
		utils.SuccessResp(w, http.StatusOK, "Data not found")
		return
	}
	w.Header().Set("Entity", fmt.Sprintf("%s", id))
	utils.SuccessResp(w, http.StatusOK, response)
}

// UpdateVendor controller handles Vendor update api

func UpdateVendor(w http.ResponseWriter, r *http.Request) {
	var vendor models.Vendor
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 451,Error occured while converting ID from hex")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	userDetails, err := GetUsers(w, r)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 458,Error occurred while getting user details")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	tenantId, err := primitive.ObjectIDFromHex(userDetails.TENANT_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 465,Error occurred while converting TENANT_ID from hex")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	businessId, err := primitive.ObjectIDFromHex(userDetails.BUSINESS_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 472,Error occurred while converting BUSINESS_ID from hex")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	businessUnitId, err := primitive.ObjectIDFromHex(userDetails.BUSINESS_UNIT_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 479,Error occured while converting BUSINESS_UNIT_ID from hex")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	userId := userDetails.USER_ID

	err = r.ParseMultipartForm(32 << 20) // parse up to 32 MB of data
	if err != nil {
		utils.ERROR(w, http.StatusUnprocessableEntity, err)
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 489,Error parsing multipart form")
		return
	}

	jsonStr := r.FormValue("json")
	err = json.Unmarshal([]byte(jsonStr), &vendor)
	if err != nil {
		utils.ERROR(w, http.StatusInternalServerError, err)
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 497,Error while unmarshalling JSON")
		return
	}
	vendor.UpdatedBy = userId

	err = vendor.Validate("update")
	if err != nil {
		utils.ERROR(w, http.StatusUnprocessableEntity, err)
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 505,Error occured while updating vendor")
		return
	}

	//loops through the vendor contact array an adds vendor contact id if it does not exist
	for x := range vendor.VendorContact {
		if vendor.VendorContact[x].VendorContactId == 0 {
			lastVendor, err := models.FindLastVendorContactByID(db.GetCollection(VendorColl), id, tenantId, businessId, businessUnitId)
			if err != nil {
				utils.ERROR(w, http.StatusInternalServerError, err)
				utils.GetLogger().WithError(err).Error("In vendorcontroller line 515,Error on finding last vendor contact by ID")
				return
			}
			vendor.VendorContact[x].VendorContactId = lastVendor + 1
		}
	}

	result, err := vendor.UpdateVendor(db.GetCollection(VendorColl), id, tenantId, businessId, businessUnitId)
	if err != nil {
		utils.ERROR(w, http.StatusInternalServerError, err)
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 525,Error in UpdateVendor")
		return
	}
	if result == "Vendor Not Found" {
		utils.FailureResp(w, http.StatusOK, "Data not found")
		return
	}

	var documentArray []models.Document
	//upload resume
	documents := r.MultipartForm.File["documents"]
	// resumeFiles := r.FormValue.File[]
	vendorInsertIdString := ExtractObjectID(fmt.Sprintf("%v", id))

	fmt.Println(len(documents))
	if len(documents) != 0 {
		CreateUniqueResumeName := CreateUniqueFileName(documents, vendorInsertIdString, emptyObjectId, "vendor_document")

		vendorDocument := models.Document{}
		//upload the resume in s3 bucket
		for i, file := range documents {
			s3Bucket.MultiDocsUpload([]*multipart.FileHeader{file}, "vendor/documents/", CreateUniqueResumeName[i])
			vendorDocument.VendorDocumentsLocation = fmt.Sprintf("vendor/documents/%s", CreateUniqueResumeName[i])
			vendorDocument.VendorDocumentsUploadName = documents[i].Filename
			vendorDocument.VendorDocumentsUniqueName = CreateUniqueResumeName[i]
			vendorDocument.VendorDocumentsStatus = true
			documentArray = append(documentArray, vendorDocument)
		}
		vendor.VendorDocuments = documentArray
		_, err = models.PushVendorDocuments(db.GetCollection(VendorColl), tenantId, businessId, businessUnitId, id, vendor)
		if err != nil {
			utils.ERROR(w, http.StatusInternalServerError, err)
			utils.GetLogger().WithError(err).Error("In vendorcontroller line 557,Error occured while pushing vendor document")
			return
		}
	}
	utils.SuccessResp(w, http.StatusOK, result)
}

func UpdateHotlistVendor(w http.ResponseWriter, r *http.Request) {
	var vendor models.Vendor
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 571,Error occurred while converting ID from hex")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	userDetails, err := GetUsers(w, r)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 578,Error occurred while getting user details")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	tenantId, err := primitive.ObjectIDFromHex(userDetails.TENANT_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 585,Error occurred while converting TENANT_ID from hex")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	businessId, err := primitive.ObjectIDFromHex(userDetails.BUSINESS_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 592,Error occurred while converting BUSINESS_ID from hex")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	businessUnitId, err := primitive.ObjectIDFromHex(userDetails.BUSINESS_UNIT_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 599,Error occured while converting BUSINESS_UNIT_ID from hex")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&vendor)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 606,Error occurred while decoding JSON from request body")
		utils.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	err = vendor.Validate("update")
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 613,Error on validating vendor for update")
		utils.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	result, err := vendor.UpdateHotlistVendor(db.GetCollection(VendorColl), id, tenantId, businessId, businessUnitId)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 620,Error occurred in UpdateHotlistVendor")
		utils.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if result == "Vendor Not Found" {
		utils.GetLogger().Error("In vendorcontroller line 625,Vendor not found in UpdateHotlistVendor")
		utils.FailureResp(w, http.StatusOK, "Data not found")
		return
	}

	utils.SuccessResp(w, http.StatusOK, result)
}

func GetAvailableDetails(w http.ResponseWriter, r *http.Request) {
	filter := models.VendorFilter{}
	cursor := filter.GetAvailableDetails(db.GetCollection(VendorColl))
	utils.SuccessResp(w, http.StatusOK, cursor)
}

// FilterVendor controller handles the Vendor filter api
// and returns the requested Vendor data
func FilterVendor(w http.ResponseWriter, r *http.Request) {
	userDetails, err := GetUsers(w, r)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 644,Error occurred while getting user details")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	tenantId, err := primitive.ObjectIDFromHex(userDetails.TENANT_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 651,Error occurred while converting TENANT_ID from hex")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	businessId, err := primitive.ObjectIDFromHex(userDetails.BUSINESS_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 658,Error occurred while converting BUSINESS_ID from hex")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	businessUnitId, err := primitive.ObjectIDFromHex(userDetails.BUSINESS_UNIT_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 665,Error occurred while converting BUSINESS_UNIT_ID from hex")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 672,Error occurred while reading request body in FilterVendor")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	filter := models.VendorFilter{}

	err = json.Unmarshal(body, &filter)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 681,Error while unmarshalling JSON in FilterVendor")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	masterGotten, err := filter.FilterVendor(db.GetCollection(VendorColl), tenantId, businessId, businessUnitId, db.GetCollection(CountryColl), db.GetDB())
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 688,Error occurred in FilterVendor")
		utils.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if masterGotten == nil {
		utils.GetLogger().Error("In vendorcontroller line 693,Data not found in FilterVendor")
		utils.FailureResp(w, http.StatusOK, "Data not found")
		return
	}
	utils.SuccessResp(w, http.StatusOK, masterGotten)
}

func FilterVendorTest(w http.ResponseWriter, r *http.Request) {
	userDetails, err := GetUsers(w, r)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 703, Error occurred while getting user details")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	tenantId, err := primitive.ObjectIDFromHex(userDetails.TENANT_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 710, Error occurred while converting TENANT_ID from hex")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	businessId, err := primitive.ObjectIDFromHex(userDetails.BUSINESS_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 717, Error occurred while converting BUSINESS_ID from hex")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	businessUnitId, err := primitive.ObjectIDFromHex(userDetails.BUSINESS_UNIT_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 724, Error occurred while converting BUSINESS_UNIT_ID from hex")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 731, Error occurred while reading request body in FilterVendor")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	filter := models.VendorFilter{}

	err = json.Unmarshal(body, &filter)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 740, Error while unmarshalling JSON in FilterVendor")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	// Validate and parse fromDate and toDate
	fromTime, toTime, err := models.ParseDateParameters(filter.FromDate, filter.ToDate)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 748,Invalid date parameters")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	// Call FilterVendorTest method with fromDate and toDate
	filteredResults, err := filter.FilterVendorTest(
		db.GetCollection(VendorColl),
		tenantId,
		businessId,
		businessUnitId,
		db.GetCollection(CountryColl),
		db.GetDB(),
		fromTime,
		toTime,
	)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 766, Error occurred in FilterVendor")
		utils.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	// Check if there are no results
	if len(filteredResults) == 0 {
		utils.GetLogger().Error("In vendorcontroller line 773, Data not found in FilterVendor")
		utils.FailureResp(w, http.StatusOK, "Data not found")
		return
	}

	utils.SuccessResp(w, http.StatusOK, filteredResults)
}

func GetFilterTechnologies(w http.ResponseWriter, r *http.Request) {

	userDetails, err := GetUsers(w, r)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 785,Error occurred while getting user details")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	tenantId, err := primitive.ObjectIDFromHex(userDetails.TENANT_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 792,Error occurred while converting TENANT_ID from hex")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	businessId, err := primitive.ObjectIDFromHex(userDetails.BUSINESS_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 799,Error occurred while converting BUSINESS_ID from hex")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	businessUnitId, err := primitive.ObjectIDFromHex(userDetails.BUSINESS_UNIT_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 806,Error occurred while converting BUSINESS_UNIT_ID from hex")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	vendor := models.Vendor{}

	allSkills := vendor.GetTechnologies(db.GetCollection(VendorColl), tenantId, businessId, businessUnitId)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 815,Error occurred while fetching skills")
		utils.ERROR(w, http.StatusInternalServerError, err)
	}
	utils.SuccessResp(w, http.StatusOK, allSkills)
}

func GetCountry(w http.ResponseWriter, r *http.Request) {
	countryList, err := models.GetCountryList(db.GetCollection(CountryColl))
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 824,Error occurred while fetching the country")
		utils.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	utils.SuccessResp(w, http.StatusOK, countryList)
}

func GetState(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 834,Error occurred while reading request body in GetState")
		utils.ERROR(w, http.StatusBadRequest, err)
	}
	country := models.Country{}
	err = json.Unmarshal(body, &country)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 840,Error while unmarshalling JSON in GetState")
		utils.ERROR(w, http.StatusBadRequest, err)
	}
	if country.CountryId == 0 {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 844,Required Country ID is missing in GetState")
		err = errors.New("required Country ID")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	stateList, err := models.GetStateList(db.GetCollection(CountryColl), country.CountryId)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 852,Error in GetStateList")
		utils.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if stateList == nil {
		utils.GetLogger().Error("In vendorcontroller line 857,Data not found in GetStateList")
		utils.FailureResp(w, http.StatusOK, "Data not found")
	} else {
		utils.SuccessResp(w, http.StatusOK, stateList)
	}
}

func GetCity(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 868,Error occurred while reading request body in GetCity")
		utils.ERROR(w, http.StatusBadRequest, err)
	}

	country := models.Country{}
	err = json.Unmarshal(body, &country)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 875,Error unmarshalling JSON in GetCity")
		utils.ERROR(w, http.StatusBadRequest, err)
	}

	if country.CountryId == 0 {
		err = errors.New("required Country ID")
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 881,Required Country ID is missing in GetCity")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	if country.StateId == 0 {
		err = errors.New("required State ID")
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 888,Required State ID is missing in GetCity")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	cityList, err := models.GetCityList(db.GetCollection(CountryColl), country.CountryId, country.StateId)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 895,Error in GetCityList")
		utils.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if cityList == nil {
		utils.GetLogger().Error("In vendorcontroller line 900,Data not found in GetCityList")
		utils.FailureResp(w, http.StatusOK, "Data not found")
	} else {
		utils.SuccessResp(w, http.StatusOK, cityList)
	}
}

func GetCountryCode(w http.ResponseWriter, r *http.Request) {

	countryCodeList, err := models.GetCountryCode(db.GetCollection(CountryColl))
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 911,Error occurred while fetching the country code")
		utils.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	utils.SuccessResp(w, http.StatusOK, countryCodeList)
}

func ShareMail(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 921,Error occurred while reading request body in ShareMail")
		utils.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	userDetails, err := GetUsers(w, r)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 927,Error occurred while getting user details in ShareMail")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	tenantId, err := primitive.ObjectIDFromHex(userDetails.TENANT_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 934,Error occurred while converting TENANT_ID from hex in ShareMail")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	businessId, err := primitive.ObjectIDFromHex(userDetails.BUSINESS_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 941,Error occurred while converting BUSINESS_ID from hex in ShareMail")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	businessUnitId, err := primitive.ObjectIDFromHex(userDetails.BUSINESS_UNIT_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 948,Error occurred while converting BUSINESS_UNIT_ID from hex in ShareMail")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}
	mail := models.Mail{}
	// json.NewDecoder(r.Body).Decode(&hotlist)
	err = json.Unmarshal(body, &mail)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 956,Error unmarshalling JSON in ShareMail")
		utils.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	// err = json.Unmarshal(body, &hotlist)
	result1, err := mail.VendorMailShare(tenantId, businessId, businessUnitId)
	// result2, err := ApplicantProfessionalDetails.FindApplicantProfessionalDetailsByID(db.GetDB(CollectionAPD), _id2, tenantId, businessId, businessUnitId)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 964,Error in VendorMailShare")
		utils.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	utils.SuccessResp(w, http.StatusOK, result1)
}

func Search(w http.ResponseWriter, r *http.Request) {
	// Vendor := models.Vendor{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 975,Error occurred while reading request body in Search")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	search := models.Search{}
	err = json.Unmarshal(body, &search)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 983,Error unmarshalling JSON in Search")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}
	if len(search.SearchValue) < 3 {
		utils.GetLogger().WithError(errors.New("in vendorcontroller line 988,required a minimum of 3 letters")).Error("Search value length less than 3 in Search")
		utils.ERROR(w, http.StatusBadRequest, errors.New("required a minimum of 3 letters"))
		return
	}
	parameters := r.URL.Query()

	userDetails, err := GetUsers(w, r)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 996,Error occurred while getting user details in Search")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	tenantId, err := primitive.ObjectIDFromHex(userDetails.TENANT_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 1003,Error occurred while converting TENANT_ID from hex in Search")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	businessId, err := primitive.ObjectIDFromHex(userDetails.BUSINESS_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 1010,Error occurred while converting BUSINESS_ID from hex in Search")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	businessUnitId, err := primitive.ObjectIDFromHex(userDetails.BUSINESS_UNIT_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 1017,Error occurred while converting BUSINESS_UNIT_ID from hex in Search")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	//results, err := Vendor.FindAllVendors(db.GetDB(VendorColl), tenantId, businessId, businessUnitId, sort, db.GetDB(CountryColl))

	sort := parameters.Get("sortBy")
	// results, err := Vendor.FindAllVendors(db.GetDB(VendorColl), tenantId, businessId, businessUnitId, sort, db.GetDB(CountryColl))
	// if err != nil {

	// 	utils.ERROR(w, http.StatusInternalServerError, err)
	// 	return
	// }
	result, err := search.VendorSearch(db.GetCollection(VendorColl), tenantId, businessId, businessUnitId, sort, db.GetCollection(CountryColl), db.GetDB())
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 1033,Error in VendorSearch")
		utils.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if result == nil {
		utils.GetLogger().Error("In vendorcontroller line 1038,No data found in Search")
		utils.FailureResp(w, http.StatusOK, "Data not found")
		return
	}
	utils.SuccessResp(w, http.StatusOK, result)
}

// GetVendorNames controller provides a list of all the available vendors
func GetVendorNames(w http.ResponseWriter, r *http.Request) {

	userDetails, err := GetUsers(w, r)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 1050,Error occurred while getting user details in GetVendorNames")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	tenantId, err := primitive.ObjectIDFromHex(userDetails.TENANT_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 1057,Error occurred while converting TENANT_ID from hex in GetVendorNames")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	businessId, err := primitive.ObjectIDFromHex(userDetails.BUSINESS_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 1064,Error occurred while converting BUSINESS_ID from hex in GetVendorNames")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	businessUnitId, err := primitive.ObjectIDFromHex(userDetails.BUSINESS_UNIT_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 1071,Error occurred while converting BUSINESS_UNIT_ID from hex in GetVendorNames")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	masterGotten, err := models.FetchVendorNames(db.GetCollection(VendorColl), tenantId, businessId, businessUnitId)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 1078,Error in FetchVendorNames")
		utils.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	utils.SuccessResp(w, http.StatusOK, masterGotten)
}

// CreateVendorContact creates a new vendor contact
func CreateVendorContact(w http.ResponseWriter, r *http.Request) {

	userDetails, err := GetUsers(w, r)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 1091,Error occurred while getting user details in CreateVendorContact")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 1098,Error occurred while reading request body in CreateVendorContact")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	vendorContact := models.VendorContactUpdate{}
	err = json.Unmarshal(body, &vendorContact)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 1106,Error occurred while unmarshalling JSON in CreateVendorContact")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	vendorContact.TenantId, err = primitive.ObjectIDFromHex(userDetails.TENANT_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 1113,Error occurred while converting TENANT_ID from hex in CreateVendorContact")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	vendorContact.BusinessId, err = primitive.ObjectIDFromHex(userDetails.BUSINESS_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 1120,Error occurred while converting BUSINESS_ID from hex in CreateVendorContact")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	vendorContact.BusinessUnitId, err = primitive.ObjectIDFromHex(userDetails.BUSINESS_UNIT_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 1127,Error occurred while converting BUSINESS_UNIT_ID from hex in CreateVendorContact")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	vendorContact.UpdatedBy = userDetails.USER_ID

	err = vendorContact.ValidateContact("create")
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 1136,Error occurred while validating vendor contact in CreateVendorContact")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	//loops through the vendor contact array an adds vendor contact id if it does not exist
	var lastVendorId int
	for x := range vendorContact.VendorContact {
		if vendorContact.VendorContact[x].VendorContactId == 0 && lastVendorId == 0 {
			countRcvd, err := models.FindVendorContactByEmail(db.GetCollection(VendorColl), vendorContact.TenantId, vendorContact.BusinessId, vendorContact.BusinessUnitId, vendorContact.VendorCompanyID, vendorContact.VendorContact[x].Email)
			if err != nil {
				utils.GetLogger().WithError(err).Error("In vendorcontroller line 1147,Error occurred while finding vendor contact by email in CreateVendorContact")
				utils.ERROR(w, http.StatusInternalServerError, err)
				return
			}
			if countRcvd == 0 {
				lastVendor, err := models.FindLastVendorContactByID(db.GetCollection(VendorColl), vendorContact.VendorCompanyID, vendorContact.TenantId, vendorContact.BusinessId, vendorContact.BusinessUnitId)
				if err != nil {
					utils.GetLogger().WithError(err).Error("In vendorcontroller line 1154,Error occurred while finding last vendor contact by ID in CreateVendorContact")
					utils.ERROR(w, http.StatusInternalServerError, err)
					return
				}
				vendorContact.VendorContact[x].VendorContactId = lastVendor + 1
				lastVendorId = lastVendor + 1
			} else {
				utils.GetLogger().Error(vendorContact.VendorContact[x].Email + " already exists in CreateVendorContact,in vendorcontroller line 1081")
				utils.FailureResp(w, http.StatusOK, vendorContact.VendorContact[x].Email+" already exists")
				return
			}

		} else if vendorContact.VendorContact[x].VendorContactId == 0 {
			countRcvd, err := models.FindVendorContactByEmail(db.GetCollection(VendorColl), vendorContact.TenantId, vendorContact.BusinessId, vendorContact.BusinessUnitId, vendorContact.VendorCompanyID, vendorContact.VendorContact[x].Email)
			if err != nil {
				utils.GetLogger().WithError(err).Error("In vendorcontroller line 1169,Error occurred while finding vendor contact by email in CreateVendorContact")
				utils.ERROR(w, http.StatusInternalServerError, err)
				return
			}
			if countRcvd == 0 {
				vendorContact.VendorContact[x].VendorContactId = lastVendorId + 1
				lastVendorId += 1
			} else {
				utils.GetLogger().Error(vendorContact.VendorContact[x].Email + " already exists in CreateVendorContact,in vendorcontroller line 1097")
				utils.FailureResp(w, http.StatusOK, vendorContact.VendorContact[x].Email+" already exists")
				return
			}
		}
	}

	err = models.UpdateContactInfoForVendor(db.GetCollection(VendorColl), vendorContact)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 1186,Error occurred while updating contact info for vendor in CreateVendorContact")
		utils.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	utils.SuccessResp(w, http.StatusOK, "vendor contact updated successfully")

}

// GetVendorContactNames gets all the vendor contact names for the provided vendor
func GetVendorContactNames(w http.ResponseWriter, r *http.Request) {

	parameters := r.URL.Query()

	userDetails, err := GetUsers(w, r)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 1202,Error occurred while getting user details in GetVendorContactNames")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	tenantId, err := primitive.ObjectIDFromHex(userDetails.TENANT_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 1209,Error occurred while converting TENANT_ID from hex in GetVendorContactNames")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	businessId, err := primitive.ObjectIDFromHex(userDetails.BUSINESS_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 1216,Error occurred while converting BUSINESS_ID from hex in GetVendorContactNames")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	businessUnitId, err := primitive.ObjectIDFromHex(userDetails.BUSINESS_UNIT_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 1223,Error occurred while converting BUSINESS_UNIT_ID from hex in GetVendorContactNames")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	vendorIdStr := parameters.Get("vendorId")
	vendorId, err := primitive.ObjectIDFromHex(vendorIdStr)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 1231,Error occurred while converting vendor ID from hex in GetVendorContactNames")
		utils.ERROR(w, http.StatusBadRequest, errors.New("vendor Id error"))
		return
	}

	masterGotten, err := models.FetchVendorContactNames(db.GetCollection(VendorColl), tenantId, businessId, businessUnitId, vendorId)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 1238,Error occurred while fetching vendor contact names in GetVendorContactNames")
		utils.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	utils.SuccessResp(w, http.StatusOK, masterGotten)
}

// api to check if vendor name exisits
func VendorNameValidate(w http.ResponseWriter, r *http.Request) {
	parameters := r.URL.Query()

	userDetails, err := GetUsers(w, r)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 1252,Error occurred while getting user details in VendorNameValidate")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	tenantId, err := primitive.ObjectIDFromHex(userDetails.TENANT_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 1259,Error occurred while converting TENANT_ID from hex in VendorNameValidate")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	businessId, err := primitive.ObjectIDFromHex(userDetails.BUSINESS_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 1266,Error occurred while converting BUSINESS_ID from hex in VendorNameValidate")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	businessUnitId, err := primitive.ObjectIDFromHex(userDetails.BUSINESS_UNIT_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 1273,Error occurred while converting BUSINESS_UNIT_ID from hex in VendorNameValidate")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	vendorCompanyName := parameters.Get("vendorCompanyName")
	if len(vendorCompanyName) == 0 {
		utils.GetLogger().Error("In vendorcontroller line 1280,Required vendor company name not provided in VendorNameValidate")
		utils.ERROR(w, http.StatusBadRequest, errors.New("required vendor company name"))
		return
	}

	vendor := models.Vendor{}

	//Check to verify if the Vendor Exists by using the Vendor Company Name
	countRcvd, err := vendor.FindVendorByName(db.GetCollection(VendorColl), vendorCompanyName, tenantId, businessId, businessUnitId)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 1290,Error occurred while checking if the vendor exists by name in VendorNameValidate")
		utils.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if countRcvd == 0 {
		utils.SuccessResp(w, http.StatusOK, "new vendor")
		return
	} else {
		utils.FailureResp(w, http.StatusOK, "vendor already exists")
		return
	}
}

// api to check if vendor contact exisits
func VendorContactValidate(w http.ResponseWriter, r *http.Request) {
	parameters := r.URL.Query()

	userDetails, err := GetUsers(w, r)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 1309,Error occurred while getting user details in VendorContactValidate")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	tenantId, err := primitive.ObjectIDFromHex(userDetails.TENANT_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 1316,Error occurred while converting TENANT_ID from hex in VendorContactValidate")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	businessId, err := primitive.ObjectIDFromHex(userDetails.BUSINESS_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 1323,Error occurred while converting BUSINESS_ID from hex in VendorContactValidate")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	businessUnitId, err := primitive.ObjectIDFromHex(userDetails.BUSINESS_UNIT_ID)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 1330,Error occurred while converting BUSINESS_UNIT_ID from hex in VendorContactValidate")
		utils.ERROR(w, http.StatusBadRequest, err)
		return
	}

	vendorIdStr := parameters.Get("vendorId")
	vendorId, err := primitive.ObjectIDFromHex(vendorIdStr)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 1338,Error occurred while converting vendor ID from hex in VendorContactValidate")
		utils.ERROR(w, http.StatusBadRequest, errors.New("vendor id error"))
		return
	}

	vendorContactEmail := parameters.Get("vendorContactEmail")
	if len(vendorContactEmail) == 0 {
		utils.GetLogger().Error("In vendorcontroller line 1345,Required vendor contact email ID not provided in VendorContactValidate")
		utils.ERROR(w, http.StatusBadRequest, errors.New("required vendor contact email id"))
		return
	}

	//Check to verify if the Vendor Exists by using the Vendor Company Name
	countRcvd, err := models.FindVendorContactByEmail(db.GetCollection(VendorColl), tenantId, businessId, businessUnitId, vendorId, vendorContactEmail)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 1353,Error occurred while checking if the vendor contact exists by email in VendorContactValidate")
		utils.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if countRcvd == 0 {
		utils.SuccessResp(w, http.StatusOK, "new vendor")
		return
	} else {
		utils.FailureResp(w, http.StatusOK, "vendor already exists")
		return
	}
}

func GetDocumentLink(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	path := v.Get("path")

	result1, err := s3Bucket.GetFileFromS3(path)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In vendorcontroller line 1372,Error occurred while getting document link from S3 in GetDocumentLink")
		utils.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	utils.SuccessResp(w, http.StatusOK, result1)
}

func MultiDelete(w http.ResponseWriter, r *http.Request) {

}
