package utils_test

import (
	"github.com/wwi21seb-projekt/server-beta/internal/utils"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestGetActivationEmailBody tests if GetActivationEmailBody returns the expected HTML content
func TestGetActivationEmailBody(t *testing.T) {
	token := "123456"
	body := utils.GetActivationEmailBody(token)
	currentYear := time.Now().Year()

	if !strings.Contains(body, token) {
		t.Errorf("Expected body to contain the token %s, but it didn't", token)
	}

	if !strings.Contains(body, strconv.Itoa(currentYear)) {
		t.Errorf("Expected body to contain the current year %d, but it didn't", currentYear)
	}

	expectedStrings := []string{
		"<!DOCTYPE html>",
		"<html lang=\"de\">",
		"<head>",
		"<meta charset=\"UTF-8\">",
		"Verifizierungscode",
		"Bitte verwenden Sie den folgenden Code, um die Registrierung Ihres Accounts bei Server Beta abzuschließen:",
		"Dieser Code ist 24 Stunden gültig.",
		"© " + strconv.Itoa(currentYear) + " Server Beta - Alle Rechte vorbehalten.",
	}

	for _, str := range expectedStrings {
		if !strings.Contains(body, str) {
			t.Errorf("Expected body to contain %s, but it didn't", str)
		}
	}
}

// TestGetWelcomeEmailBody tests if GetWelcomeEmailBody returns the expected HTML content with the username
func TestGetWelcomeEmailBody(t *testing.T) {
	username := "testuser"
	body := utils.GetWelcomeEmailBody(username)
	currentYear := time.Now().Year()

	if !strings.Contains(body, username) {
		t.Errorf("Expected body to contain the username %s, but it didn't", username)
	}

	if !strings.Contains(body, strconv.Itoa(currentYear)) {
		t.Errorf("Expected body to contain the current year %d, but it didn't", currentYear)
	}

	expectedStrings := []string{
		"<!DOCTYPE html>",
		"<html lang=\"de\">",
		"<head>",
		"<meta charset=\"UTF-8\">",
		"Willkommen bei Server Beta!",
		"Hallo " + username + "!",
		"Dein Account wurde erfolgreich verifiziert.",
		"© " + strconv.Itoa(currentYear) + " Server Beta - Alle Rechte vorbehalten.",
	}

	for _, str := range expectedStrings {
		if !strings.Contains(body, str) {
			t.Errorf("Expected body to contain %s, but it didn't", str)
		}
	}
}
