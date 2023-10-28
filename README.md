# OpenAI Proxy

OpenAI Proxy is a powerful proxy and management panel designed to simplify the management of your OpenAI endpoints in one convenient location. This project is powered by the Go-Admin panel, making it user-friendly and efficient. You can access the Go-Admin panel [here](https://github.com/GoAdminGroup/go-admin).

## Getting Started

Follow these steps to set up and run OpenAI Proxy:

1. **Ensure you have the `admin.db` file in your repository.**

2. **Modify the configuration file in the root directory.**
   The default configuration is provided below. You can change the settings in the `config.json` file according to your requirements.

   ```json
   {
       "debug": true,
       "index": "/",
       "prefix": "admin",
       "language": "en",
       "database": {
           "default": {
               "host": "admin.db",
               "max_idle_con": 50,
               "max_open_con": 150,
               "driver": "sqlite"
           }
       },
       "store": {
           "path": "./uploads",
           "prefix": "uploads"
       },
       "access_log": "./logs/access.log",
       "error_log": "./logs/error.log",
       "info_log": "./logs/info.log",
       "access_assets_log_off": true,
       "theme": "adminlte",
       "bootstrap_file_path": "./bootstrap.go",
       "color_scheme": "skin-black"
   }
   ```

3. **Run the application:**
   - To run OpenAI Proxy using Docker, use the following command:
     ```bash
     docker-compose up
     ```
   - To run it with Go, use the following command:
     ```bash
     go run main.go
     ```
   - You can also build the application as needed.

4. **Initialization:**
   Place the required JSON configuration files (`endpoints.json`, `models.json`, `user.json`) in the `config` folder in the root repository. For example, a `models.json` file might look like this:

   ```json
   {
       "models": [
           {
               "name": "openai",
               "sub_models": [
                   {
                       "name": "whisper-1"
                   },
                   {
                       "name": "gpt-4"
                   },
                   {
                       "name": "gpt-4-0613"
                   },
                   // ... (other models)
               ]
           }
       ]
   }
   ```

   Additional JSON configurations will be added in the future. Stay tuned for updates.

5. **Customize Ports and Database:**
   By default, the management panel is on port 8081, and the proxy is on port 8080. You can customize these settings by setting environment variables:

   - `PORT` to specify the proxy port.
   - `PANEL_PORT` to specify the management panel port.
   - `DB_URL` and `DB_NAME` to configure the database variables.

## Additional Configuration

Additional configuration options and features will be added to OpenAI Proxy in future updates. Stay tuned for more information.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
