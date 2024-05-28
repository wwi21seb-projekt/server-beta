package utils

import (
	"fmt"
	"time"
)

// GetActivationEmailBody returns the HTML body for an activation email sending the activation code
func GetActivationEmailBody(token string) string {
	currentYear := time.Now().Year()
	return fmt.Sprintf(`
	<!DOCTYPE html>
	<html lang="de">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<style>
			body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
			.container { width: 100%%; max-width: 600px; margin: auto; background-color: #f9f9f9; padding: 20px; }
			.header { background-color: #007bff; color: white; padding: 10px 20px; text-align: center; }
			.content { margin: 20px; text-align: center; }
			.code { font-size: 24px; color: #007bff; padding: 20px; margin: 20px 0; background-color: #eef; border-radius: 8px; display: inline-block; }
			.footer { font-size: 0.8em; text-align: center; margin-top: 20px; color: #666; }
		</style>
	</head>
	<body>
		<div class="container">
			<div class="header">
				Verifizierungscode
			</div>
			<div class="content">
				<p>Hallo!</p>
				<p>Bitte verwenden Sie den folgenden Code, um die Registrierung Ihres Accounts bei Server Beta abzuschließen:</p>
				<div class="code">%s</div>
				<p>Dieser Code ist 2 Stunden gültig. Geben Sie diesen Code auf der entsprechenden Seite zu Ihrer Registrierung ein.</p>
			</div>
			<div class="footer">
				© %d Server Beta - Alle Rechte vorbehalten.
				<br>
				Mehr Informationen finden Sie in unserem <a href="https://server-beta.de/api/imprint">Impressum</a>.
			</div>
		</div>
	</body>
	</html>`, token, currentYear)
}

// GetWelcomeEmailBody returns the HTML body for a welcome email
func GetWelcomeEmailBody(username string) string {
	currentYear := time.Now().Year()
	return fmt.Sprintf(`
	<!DOCTYPE html>
	<html lang="de">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<style>
			body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
			.container { width: 100%%; max-width: 600px; margin: auto; background-color: #f9f9f9; padding: 20px; }
			.header { background-color: #007bff; color: white; padding: 10px 20px; text-align: center; }
			.content { margin: 20px; text-align: center; }
			.footer { font-size: 0.8em; text-align: center; margin-top: 20px; color: #666; }
		</style>
	</head>
	<body>
		<div class="container">
			<div class="header">
				Willkommen bei Server Beta!
			</div>
			<div class="content">
				<p>Hallo %s!</p>
				<p>Let's go!</p>
				<p>Dein Account wurde erfolgreich verifiziert. Du kannst jetzt unser Netzwerk nutzen.</p>
				<p>Wir laden dich ein, aktiv Teil unserer Community zu werden und uns deine Gedanken mitzuteilen.</p>
			</div>
			<div class="footer">
				© %d Server Beta - Alle Rechte vorbehalten.
				<br>
				Mehr Informationen finden Sie in unserem <a href="https://server-beta.de/api/imprint">Impressum</a>.
			</div>
		</div>
	</body>
	</html>`, username, currentYear)
}

// GetPasswordResetEmailBody returns the HTML body for a password reset email
func GetPasswordResetEmailBody(username string, resetToken string) string {
	currentYear := time.Now().Year()
	return fmt.Sprintf(`
	<!DOCTYPE html>
	<html lang="de">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<style>
			body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
			.container { width: 100%%; max-width: 600px; margin: auto; background-color: #f9f9f9; padding: 20px; }
			.header { background-color: #007bff; color: white; padding: 10px 20px; text-align: center; }
			.content { margin: 20px; text-align: center; }
			.code { font-size: 24px; color: #007bff; padding: 20px; margin: 20px 0; background-color: #eef; border-radius: 8px; display: inline-block; }
			.footer { font-size: 0.8em; text-align: center; margin-top: 20px; color: #666; }
		</style>
	</head>
	<body>
		<div class="container">
			<div class="header">
				Passwort zurücksetzen
			</div>
			<div class="content">
				<p>Hallo %s!</p>
				<p>Ihr Code lautet:</p>
				<div class="code">%s</div>
				<p>Verwenden Sie diesen Code, um Ihr Passwort zurückzusetzen.</p>
			</div>
			<div class="footer">
				© %d Server Beta - Alle Rechte vorbehalten.
				<br>
				Mehr Informationen finden Sie in unserem <a href="https://server-beta.de/api/imprint">Impressum</a>.
			</div>
		</div>
	</body>
	</html>`, username, resetToken, currentYear)
}
