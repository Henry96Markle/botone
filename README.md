# botone

Clone the project, and setup a ```config.env``` file with the following variables:

- ```TOKEN``` -> Your bot's token
- ```OWNER``` -> Your Telegram ID
- ```PORT``` -> Port for listening to webhook updates; (Just set it to 80)
- ```CONNECTION_STRING``` -> Your MongoDB cluster connection string
- ```LOGGING_TO_CHANNEL``` -> It's a boolean; decide whether you want use a channel for logging or not
- ```LOG_CHANNEL_ID``` -> The ID of that channel; remember to add your bot to the channel

Initialize it as a Git repository and add the .env file to ```.gitignore```.
Finally, create a Heroku app and push the source code to the master branch, remotely.<br>
Manually setup the environment variables in Heroku, if needed.