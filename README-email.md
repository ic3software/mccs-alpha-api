# MCCS Emails

MCCS will send emails to users for various reasons, either to a user's email address or to the email address that has been associated with an entity.

Admins can also receive emails from MCCS which are triggered by certain events.

## Types of Emails

Email Type | Sent To | Description
--- | --- | ---
[Welcome message](/template/email/welcome.html) | Entity email |  A welcome message that is sent immediately after submitting user/entity data to the `POST /signup` endpoint.
[Daily match notification](/template/email/dailyEmail.html) | Entity email | If the entity has the `receiveDailyMatchNotificationEmail` flag on, any new matches to that entity's offers and wants tags posted by other entities are sent in a summary email. The front end app needs to handle the receipt of the code in the URL to call the API to retrieve the matched entities and display them to the user.
[Trade contact](/internal/pkg/email/email.go) | Entity email | Any entity listed in the directory can contact another entity by email. MCCS sends an email to the receiving entity that reveals the sending entity's email address, so the receiver can reply directly to the sender to continue the conversation. The sender does not ever see the receiver's email address unless the receiver decides to reply.
[Transfer initiated](/internal/pkg/email/transfer.go) | Entity email | An entity who is the recipient of a transfer initiated by another entity will receive this email notifying them of the need to accept or reject the transfer. The front end app needs to handle retrieving the list of pending transfers and displaying them to the user.
[Transfer accepted](/internal/pkg/email/transfer.go) | Entity email | An email sent to the initiator of a transfer once it has been accepted by the receiver.
[Transfer rejected](/internal/pkg/email/transfer.go) | Entity email | An email sent to the initiator of a transfer once it has been rejected by the receiver.
[Transfer cancelled](/internal/pkg/email/transfer.go) | Entity email | An email sent to the receiver of a transfer once it has been cancelled by the initiator (and before the receiver has a chance to accept or reject it).
[Transfer cancelled by system](/internal/pkg/email/transfer.go) | Entity email | Entity email | An email sent to the initiator of a transfer once it has been rejected by MCCS. The usual reason this will happen is because the initiator's and/or receiver's balance will breach the maximum positive and/or negative balance limits if the transfer were to be completed.
[Password reset](/internal/pkg/email/email.go) | User or admin email | Users and admins can request a reset of their password when they forgot it. A coded URL is sent by email that the user or admin can click on to start the reset process. The front end app needs to handle the receipt of the code in the URL and initiate through the API the password reset with the new password.
[Signup notification](/internal/pkg/email/email.go) | Admin email | An email is sent to admins whenever a new user signs up in MCCS.

## Email Environment Variables

There are a few email-related environment variables in the [example configuration file](/configs/development-example.yaml):

```
daily_email_schedule: "* * 1 * * *"
balance_check_schedule: "* * 1 * * *"

receive_email:
  trade_contact_emails: true
  signup_notifications: true

sendgrid:
  key: xxx
  sender_email: xxx
```

- `daily_email_schedule` - The time when the daily match notification emails are sent
- `balance_check_schedule` - The time to run a routine that totals all transactions in the PostgreSQL database's `postings` table to ensure they add up to zero (debits and credits are equal, which is an important accounting principle in a mutual credit system). If they do not balance out at zero, an email is sent to admins informing them of the discrepancy.
- `trade_contact_emails` - If set to true, admins will receive a copy of any trade contact emails initiated by an entity. If set to true, the front end app should make this clear to entity's initiating contact.
- `signup_notifications` - If set to true, admins will receive signup notification emails.
- `sendgrid: key` - The API key provided by Sendgrid when you create an account with them.
- `sendgrid: sender_email` - The email address you want to show on all emails sent by MCCS (e.g., `support@your.org`). Admin notification and alert emails are also sent to this address by MCCS.
