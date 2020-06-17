# MCCS Emails

MCCS will send emails to users for various reasons, either to a user's email address or to the email address that has been associated with an entity.

Admins can also receive emails from MCCS which are triggered by certain events.

## Types of Emails

Email Type | Sent To | Description
--- | --- | ---
[Welcome message](#welcome-message) | Entity email |  A welcome message that is sent immediately after submitting user/entity data to the `POST /signup` endpoint.
[Daily match notification](#daily-match-notification) | Entity email | If the entity has the `receiveDailyMatchNotificationEmail` flag on, any new matches to that entity's offers and wants tags posted by other entities are sent in a summary email. The front end app needs to handle the receipt of the query parameters in the URL to call the API to retrieve the matched entities and display them to the user.
[Trade contact](#trade-contact) | Entity email | Any entity listed in the directory can contact another entity by email. MCCS sends an email to the receiving entity that reveals the sending entity's email address, so the receiver can reply directly to the sender to continue the conversation. The sender does not ever see the receiver's email address unless the receiver decides to reply.
[Transfer initiated](#transfer-initiated) | Entity email | An entity who is the recipient of a transfer initiated by another entity will receive this email notifying them of the need to accept or reject the transfer. The front end app needs to handle retrieving the list of pending transfers and displaying them to the user.
[Transfer accepted](#transfer-accepted) | Entity email | An email sent to the initiator of a transfer once it has been accepted by the receiver.
[Transfer rejected](#transfer-rejected) | Entity email | An email sent to the initiator of a transfer once it has been rejected by the receiver.
[Transfer cancelled](#transfer-cancelled) | Entity email | An email sent to the receiver of a transfer once it has been cancelled by the initiator (and before the receiver has a chance to accept or reject it).
[Transfer cancelled by system](#transfer-cancelled-by-system) | Entity email | An email sent to the initiator of a transfer once it has been rejected by MCCS. The usual reason this will happen is because the initiator's and/or receiver's balance will breach the maximum positive and/or negative balance limits if the transfer were to be completed.
[User password reset](#user-password-reset) | User email | Users can request a reset of their password when they forgot it. A URL with a unique code (an authentication token in essence) in the path parameter is sent by email to start the reset process. The front end app needs to handle the receipt of the code in the path parameter and initiate through the API the password reset with the new password and passing the unique code.
[Admin password reset](#admin-password-reset) | Admin email | See the **User password reset** description above.
[Signup notification](#signup-notification) | Admin email | An email is sent to admins whenever a new user signs up in MCCS.
[Non-zero balance notification](#non-zero-balance-notification) | Admin email | An email is sent to admins when a discrepancy is found after running a routine that totals all transactions in the PostgreSQL database's `postings` table to ensure they add up to zero (debits and credits are equal, which is an important accounting principle in a mutual credit system).

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
  template_id:
    welcome_message: xxx
    daily_match_notification: xxx
    trade_contact: xxx
    transfer_initiated: xxx
    transfer_accepted: xxx
    transfer_rejected: xxx
    transfer_cancelled: xxx
    transfer_cancelled_by_system: xxx
    user_password_reset: xxx
    admin_password_reset: xxx
    signup_notification: xxx
    non_zero_balance_notification: xxx

```

- `daily_email_schedule` - The time when the daily match notification emails are sent.
- `balance_check_schedule` - The time to run the balance checking routine to ensure all entries in the `postings` table add up to zero (should be run at least once per day).
- `trade_contact_emails` - If set to true, admins will receive a copy of any trade contact emails initiated by an entity. If set to true, the front end app should make this clear to entity's initiating contact.
- `signup_notifications` - If set to true, admins will receive signup notification emails.
- `sendgrid: key` - The API key provided by Sendgrid when you create an account with them.
- `sendgrid: sender_email` - The email address you want to show on all emails sent by MCCS (e.g., `support@your.org`). Admin notification and alert emails are also sent to this address by MCCS.
- `sendgrid: template_id` - The 12 template IDs assigned by Sendgrid to the email templates you created for each of the system-generated emails sent by MCCS.

## Sendgrid Email Templates

Emails are sent by passing the relevant data to Sendgrid who then generates the emails using pre-defined templates.

**How to create a dynamic template using SendGrid?**

1. Login to the [SendGrid dashboard](https://mc.sendgrid.com)
2. Go to Menu > Email API > [Dynamic Templates](https://mc.sendgrid.com/dynamic-templates)
3. Create a template using [Handlebars](https://handlebarsjs.com/)
4. Add the template ID in the corresponding `sendgrid: template_id` [environment variable](#email-environment-variables)

The basic templates listed below can be used as starters on top of which you can generate more customized emails. The items in curly braces (e.g., `{{entityName}}`) are provided to Sendgrid by MCCS and will be inserted automatically into the email.

### Welcome message

```
Subject: Welcome to the OCN directory!

<html>
<head>
  <title>Open Credit Network</title>
</head>
<body>
    <p>Hi {{entityName}},</p>
    <p>Thanks for signing up to the OCN directory! Your details will be reviewed by the OCN team and if everything is OK, your directory entry will go live. We will be in touch again very soon.</p>
    <p><b>The more detail you add about the goods and services you provide and are looking for, the more potential entity it will generate</b>, so please sign in to <a href="https://trade.opencredit.network/account">complete your profile</a>.</p>
    <p>If you would like to trade with another entity using Mutual Credit you will need to <a href="https://trade.opencredit.network/member-signup">apply to become a Trading Member</a>.</p>
    <p>If you know people from other entities who may be interested in being part of the Open Credit Network please <a href="https://opencredit.network/spread-the-word/">spread the word</a>.</p>
    <p>If you have any questions just let us know by replying to this email.</p>
    <p>
        In Mutuality!
        <br />
        The OCN Team
    </p>
</body>
</html>
```

### Daily match notification

```
Subject: Potential trades via the Open Credit Network

<html>
<head>
  <title>Potential trades via the Open Credit Network</title>
</head>
<body>
  <div>
    <p>Good news!</p>
    {{#if matchedOffers}}
        <h2>Matched Offers</h2>
        <p>There are new customers on the Open Credit Network who want what you're offering:</p>
        <div>
            <ul>
                {{#each matchedOffers as |value key|}}
                    <li>
                        <span style="margin-top:5px;margin-bottom:5px;margin:auto;padding: 0px 5px;display: inline-block;vertical-align: middle;border-radius: 5px;background: #003fa7;color: #ffffff;">{{ key }}</span> -
                        <a href="{{../url}}/new-matches?page=1&wants={{key}}&tagged_since={{../lastNotificationSentDate}}">
                            See the full details and get in touch
                        </a>
                    </li>
                {{/each}}
            </ul>
        </div>
    {{/if}}
    {{#if matchedWants}}
        <h2>Matched Wants</h2>
        <p>There are new suppliers on the Open Credit Network who are offering what you want::</p>
        <div>
            <ul>
                {{#each matchedWants as |value key|}}
                    <li>
                        <span style="margin-top:5px;margin-bottom:5px;margin:auto;padding: 0px 5px;display: inline-block;vertical-align: middle;border-radius: 5px;background: #f7a502;color: #ffffff;">{{ key }}</span> -
                        <a href="{{../url}}/new-matches?page=1&offers={{key}}&tagged_since={{../lastNotificationSentDate}}">
                            See the full details and get in touch
                        </a>
                    </li>
                {{/each}}
            </ul>
        </div>
    {{/if}}
  </div>
  
    <p>If your wants, offers or entity details have changed, please <a href="{{url}}/account">update your listing</a>.</p>

    <p>You can also unsubscribe from these emails by <a href="{{url}}/account">following this link</a>.</p>

    <p>Happy trading!<br>
    The Open Credit Network</p>
</body>
</html>
```

### Trade contact

```
Subject: Contact from OCN directory member

<html>
<head>
  <title></title>
</head>
<body>
    {{body}}
</body>
</html>
```

### Transfer initiated

```
Subject: OCN Transaction Requiring Your Approval

<html>
<head>
  <title></title>
</head>
<body>
    {{initiatorEntityName}} wants to {{action}}. <a href="{{url}}">Click here to review this pending transaction</a>.
</body>
</html>
```

### Transfer accepted

```
Subject: OCN Transaction Accepted

<html>
<head>
  <title></title>
</head>
<body>
  {{receiverEntityName}} has accepted the transaction you initiated for {{transferDirection}} {{amount}} Credits.
</body>
</html>
```

### Transfer rejected

```
Subject: OCN Transaction Rejected

<html>
<head>
  <title></title>
</head>
<body>
  {{receiverEntityName}} has rejected the transaction you initiated for {{transferDirection}} {{amount}} Credits.
  {{#if reason}}
    <div>Reason: {{reason}}</div>
  {{/if}}
</body>
</html>
```

### Transfer cancelled

```
Subject: OCN Transaction Cancelled

<html>
<head>
  <title></title>
</head>
<body>
  {{initiatorEntityName}} has cancelled the transaction it initiated for {{transferDirection}} {{amount}} Credits.
</body>
</html>
```

### Transfer cancelled by system

```
Subject: OCN Transaction Cancelled

<html>
<head>
  <title></title>
</head>
<body>
  The system has cancelled the transaction you initiated with {{receiverEntityName}} for the following reason: {{reason}}
</body>
</html>
```

### User password reset

```
Subject: Password Reset

<html>
<head>
  <title></title>
</head>
<body>
    Your password reset link is: <a href="{{serverAddress}}/password-reset/{{token}}">{{serverAddress}}/password-reset/{{token}}</a>
</body>
</html>
```

### Admin password reset

```
Subject: Password Reset

<html>
<head>
  <title></title>
</head>
<body>
    Your password reset link is: <a href="{{serverAddress}}/admin/password-reset/{{token}}">{{serverAddress}}/admin/password-reset/{{token}}</a>
</body>
</html>
```

### Signup notification

```
Subject: A new entity has been signed up!

<html>
<head>
  <title></title>
</head>
<body>
    Entity Name: {{entityName}}, Contact Email: {{contactEmail}}
</body>
</html>
```

### Non-zero balance notification

```
Subject: [System Check] Non-zero balance encountered

<html>
<head>
  <title></title>
</head>
<body>
  Non-zero balance encountered! Please check the timespan from {{fromTime}} to {{toTime}} in the posting table.
</body>
</html>
```
