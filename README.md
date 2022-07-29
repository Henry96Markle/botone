# botone

Clone the project, initialize it as a Git repository, and create a Heroku app and push the source code to the master branch, remotely.<br>
Manually setup the environment variables in Heroku as follows:

- ```TOKEN``` -> Your bot's token
- ```OWNER``` -> Your Telegram ID
- ```PORT``` -> Port for listening to webhook updates; (Just set it to 80)
- ```CONNECTION_STRING``` -> Your MongoDB cluster connection string
- ```LOGGING_TO_CHAT``` -> It's a boolean; decide whether you want use a channel for logging or not
- ```LOG_CHAT_ID``` -> The ID of that channel; remember to add your bot to the channel
