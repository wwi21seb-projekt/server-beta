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
	<html lang="en">
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
				Verification Code
			</div>
			<div class="content">
				<p>Hello!</p>
				<p>Please use the following code to complete your account registration at Server Beta:</p>
				<div class="code">%s</div>
				<p>This code is valid for 2 hours. Enter this code on the appropriate page for your registration.</p>
			</div>
			<div class="footer">
				© %d Server Beta - All rights reserved.
				<br>
				For more information, see our <a href="https://server-beta.de/api/imprint">imprint</a>.
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
	<html lang="en">
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
				Welcome to Server Beta!
			</div>
			<div class="content">
				<p>Hello %s!</p>
				<p>Let's go!</p>
				<p>Your account has been successfully verified. You can now use our network.</p>
				<p>We invite you to actively participate in our community and share your thoughts with us.</p>
			</div>
			<div class="footer">
				© %d Server Beta - All rights reserved.
				<br>
				For more information, see our <a href="https://server-beta.de/api/imprint">imprint</a>.
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
	<html lang="en">
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
				Password Reset
			</div>
			<div class="content">
				<p>Hello %s!</p>
				<p>Your code is:</p>
				<div class="code">%s</div>
				<p>Use this code to reset your password.</p>
			</div>
			<div class="footer">
				© %d Server Beta - All rights reserved.
				<br>
				For more information, see our <a href="https://server-beta.de/api/imprint">imprint</a>.
			</div>
		</div>
	</body>
	</html>`, username, resetToken, currentYear)
}
