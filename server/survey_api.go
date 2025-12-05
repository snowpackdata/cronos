package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/snowpackdata/cronos"
)

func (a *App) sendSlackNotification(message map[string]string, webhookURL string) {
	jsonMessage, _ := json.Marshal(message)
	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonMessage))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending Slack message: %s", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Printf("Slack notification sent successfully.")
	} else {
		log.Printf("Failed to send Slack notification. Status: %d", resp.StatusCode)
	}
}

func (a *App) alertOnSurveyCompletion(surveyID uint) (string, error) {
	// Retrieve the survey (and included questions)
	var survey cronos.Survey
	result := a.cronosApp.DB.Preload("SurveyResponses").First(&survey, surveyID)
	if result.Error != nil {
		log.Printf("Error retrieving survey: %s", result.Error)
		return "", result.Error
	}

	// Prepare the Slack message: it should contain the survey ID and all of its relevant, non-deleted questions
	var messageText string
	messageText = fmt.Sprintf("<!channel> A new survey has been submitted: (*Survey ID: %d*)\n*Responder:* %s\n*Company:* %s\n*Role:* %s\n\n",
		survey.ID, survey.UserEmail, survey.CompanyName, survey.UserRole)

	// Append each survey response to the message
	for _, response := range survey.SurveyResponses {
		formattedResponse := fmt.Sprintf("*Question %d:* %s\n- *Response*: %s\n- *Freeform Response*: %s\n\n",
			response.Step, response.Question, response.StructuredAnswer, response.FreeformAnswer)
		messageText += formattedResponse
	}

	// Send the message to Slack
	message := map[string]string{
		"text": messageText,
	}

	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if webhookURL == "" {
		return "", errors.New("no Slack webhook URL provided")
	}

	a.sendSlackNotification(message, webhookURL)

	return messageText, nil
}

// HubSpot-related functions:

// Function to convert markdown-like text to HTML
func markdownToHTML(markdown string) string {
	html := markdown
	// Remove bold (asterisk) notation
	html = strings.ReplaceAll(html, "*", "")
	// Convert newlines to <br> for HTML
	html = strings.ReplaceAll(html, "\n", "<br>")
	return html
}

// Utility function to make HubSpot API requests
func (a *App) hubSpotAPIRequest(method, url string, payload interface{}) (*http.Response, error) {
	HubSpotAPIToken := os.Getenv("HUBSPOT_API_KEY")
	if HubSpotAPIToken == "" {
		log.Println("HubSpot API token not set")
		return nil, errors.New("HubSpot API token not set")
	}

	// Convert the payload to JSON
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create a new HTTP request with the Bearer token in the header
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payloadJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", HubSpotAPIToken))

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to HubSpot: %w", err)
	}

	return resp, nil
}

// Function to create a HubSpot contact and return the Contact ID (VID)
func (a *App) createHubSpotContact(email, userRole string) (int, error) {
	HubSpotAPIURL := "https://api.hubapi.com/contacts/v1/contact"

	// Contact data to be sent
	contactData := map[string]interface{}{
		"properties": []map[string]string{
			{"property": "email", "value": email},
			{"property": "jobtitle", "value": userRole},
		},
	}

	// Make API request using the utility function
	resp, err := a.hubSpotAPIRequest("POST", HubSpotAPIURL, contactData)
	if err != nil {
		return 0, fmt.Errorf("failed to create contact in HubSpot: %w", err)
	}
	defer resp.Body.Close()

	// Check for successful response
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Printf("Failed to create HubSpot contact: Status %d", resp.StatusCode)
		return 0, errors.New("failed to create contact in HubSpot")
	}

	// Parse the response to get the Contact ID (VID)
	var contactResponse struct {
		VID int `json:"vid"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&contactResponse); err != nil {
		return 0, fmt.Errorf("failed to parse contact response: %w", err)
	}

	log.Printf("HubSpot contact created successfully, VID: %d", contactResponse.VID)
	return contactResponse.VID, nil
}

// Function to create a note for a HubSpot contact
func (a *App) createHubSpotNote(contactID int, messageText string) error {
	HubSpotAPIURL := "https://api.hubapi.com/crm/v3/objects/notes"

	// Convert markdown-like messageText to HTML
	htmlMessageText := markdownToHTML(messageText)

	// Get the current time as a Unix timestamp in milliseconds
	timestampMS := time.Now().UnixNano() / int64(time.Millisecond)

	// Note data to be sent
	noteRequest := map[string]interface{}{
		"properties": map[string]string{
			"hs_timestamp": strconv.FormatInt(timestampMS, 10), // Use Unix timestamp in milliseconds
			"hs_note_body": htmlMessageText,                    // Use the HTML version of messageText
		},
		"associations": []map[string]interface{}{
			{
				"to": map[string]interface{}{
					"id": contactID,
				},
				"types": []map[string]interface{}{
					{
						"associationCategory": "HUBSPOT_DEFINED",
						"associationTypeId":   202,
					},
				},
			},
		},
	}

	// Make API request using the utility function
	resp, err := a.hubSpotAPIRequest("POST", HubSpotAPIURL, noteRequest)
	if err != nil {
		return fmt.Errorf("failed to create note in HubSpot: %w", err)
	}
	defer resp.Body.Close()

	// Check for successful response
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Printf("Failed to create HubSpot note: Status %d", resp.StatusCode)
		return errors.New("failed to create note in HubSpot")
	}

	log.Printf("Note successfully created for HubSpot contact ID: %d", contactID)
	return nil
}

// SurveyUpsert
// Create an UPSERT API for generating a new survey from form data on the front end
// The API should be a POST request to /api/survey
// The API should accept a JSON object with the following fields:
// - survey_name (string)
// - user_email (string)
// - user_role (string)
// - company_name (string)
func (a *App) SurveyUpsert(w http.ResponseWriter, r *http.Request) {
	// Parse the JSON request body
	// Create a new survey in the database
	// Return the survey ID

	var survey cronos.Survey
	survey.SurveyType = r.FormValue("survey_type")
	survey.UserEmail = r.FormValue("user_email")
	// TODO: Check if they've already submitted a survey w/ this email

	survey.UserRole = r.FormValue("user_role")
	survey.CompanyName = r.FormValue("company_name")
	a.cronosApp.DB.Create(&survey)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(survey)
}

func (a *App) SurveyResponse(w http.ResponseWriter, r *http.Request) {
	// Parse the JSON request body
	// Retrieve the survey from the database
	// Return a 200
	vars := mux.Vars(r)
	surveyIdStr := vars["id"]
	surveyId, err := strconv.Atoi(surveyIdStr)
	if err != nil {
		http.Error(w, "Invalid survey ID", http.StatusBadRequest)
		return
	}
	// Input validation for "step"
	stepInt, err := strconv.Atoi(r.FormValue("step"))
	if err != nil {
		http.Error(w, "Invalid step value", http.StatusBadRequest)
		return
	}
	if stepInt < 1 || stepInt > 15 {
		http.Error(w, "Step value out of range", http.StatusBadRequest)
		return
	}

	var surveyResponse cronos.SurveyResponse

	surveyResponse.SurveyID = uint(surveyId)
	surveyResponse.Step = stepInt
	surveyResponse.Question = r.FormValue("question")
	surveyResponse.StructuredAnswer = r.FormValue("structured_answer")
	surveyResponse.AnswerType = r.FormValue("answer_type")
	surveyResponse.FreeformAnswer = r.FormValue("unstructured_answer")

	// Begin a transaction
	tx := a.cronosApp.DB.Begin()
	if tx.Error != nil {
		log.Printf("Error starting transaction: %s", tx.Error)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check if a response already exists for a given question (step) and survey_id
	// If so, delete the previous record
	var existingSurveyResponse cronos.SurveyResponse
	tx.Where("survey_id = ? AND step = ?", surveyResponse.SurveyID, surveyResponse.Step).First(&existingSurveyResponse)
	if existingSurveyResponse.ID != 0 {
		tx.Delete(&existingSurveyResponse)
	}

	// Save the survey response
	if err := tx.Create(&surveyResponse).Error; err != nil {
		log.Printf("Error saving survey response: %s", tx.Error)
		tx.Rollback() // Rollback transaction if there's an error
		http.Error(w, "Error saving survey response", http.StatusInternalServerError)
		return
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing transaction: %s", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// If the survey is completed, update the survey to reflect that
	// Send confirmation email, Slack alert, and create HubSpot contact
	if r.FormValue("completed") == "true" {
		var survey cronos.Survey
		a.cronosApp.DB.First(&survey, surveyId)
		survey.Completed = true
		a.cronosApp.DB.Save(&survey)

		// Send confirmation email
		if err := a.cronosApp.EmailFromAdmin(cronos.EmailTypeSurveyConfirmation, survey.UserEmail); err != nil {
			log.Printf("Error sending email confirmation. Status: %s", err)
			return
		}

		// Send Slack alert and get the message text for the note
		messageText, err := a.alertOnSurveyCompletion(survey.ID)
		if err != nil {
			log.Printf("Failed at Slack alert step on survey completion")
			return
		}

		// Call HubSpot API to create a contact
		contactID, err := a.createHubSpotContact(survey.UserEmail, survey.UserRole)
		if err != nil {
			log.Printf("Failed to create HubSpot contact: %s", err)
			return
		}

		// Call HubSpot API to create a note for the created contact
		if err := a.createHubSpotNote(contactID, messageText); err != nil {
			log.Printf("Failed to create HubSpot note: %s", err)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(surveyResponse)
}
