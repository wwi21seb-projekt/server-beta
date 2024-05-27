package utils

import (
	"fmt"
	"time"
)

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
				<p>Dieser Code ist 24 Stunden gültig. Geben Sie diesen Code auf der entsprechenden Seite Ihrer Account-Einstellungen ein.</p>
			</div>
			<div class="footer">
				© %d Server Beta - Alle Rechte vorbehalten.
				<br>
				Besuchen Sie unsere <a href="https://serverbeta.com/privacy">Datenschutzrichtlinie</a>
			</div>
		</div>
	</body>
	</html>`, token, currentYear)
}

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
				Besuchen Sie unsere <a href="https://serverbeta.com/privacy">Datenschutzrichtlinie</a>
			</div>
		</div>
	</body>
	</html>`, username, currentYear)
}
