# mattermost-webhook-interactive-test

## About
A very simple demo to send message to Mattermost via API and handle Mattermost interactive button callback

## Prerequisites
- Mattermost server

## Steps
1. Create a webhook on Mattermost
2. Replace `mattermostAddr` with your webhook URL `http://{your-mattermost-url}/hooks/{webhook-token}`
3. Replace the URL in message to point to your service URL

## Debugging
1. If there's no update after clicking on the button, go to Mattermost > System Console > Server logs and check the logs for errors (System Admin required)
2. You may need to update Mattermost server configuration to allow untrusted internal connection to your service. For more information, refer to https://docs.mattermost.com/administration/config-settings.html#allow-untrusted-internal-connections-to
