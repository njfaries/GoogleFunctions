using Google.Cloud.Functions.Framework;
using Microsoft.AspNetCore.Http;
using Microsoft.Extensions.Logging;
using System.IO;
using System.Text.Json;
using System.Threading.Tasks;

namespace SimpleHttpFunction
{
    public class Function : IHttpFunction
    {
        private readonly ILogger _logger;

        public Function(ILogger<Function> logger) =>
            _logger = logger;

        public async Task HandleAsync(HttpContext context)
        {
            HttpRequest request = context.Request;
            // Check URL parameters for "message" field
            // string message = request.Query["message"];
            string message = "";
            string buildNumber = request.Query["buildNumber"];
            string orgId = request.Query["orgForeignKey"];
            string projectId = request.Query["projectGuid"];

            _logger.LogInformation($"Build number: {buildNumber}\nOrg Foreign Key: {orgId}\nProject Id: {projectId}");

            // _logger.LogInformation(request.ToString());

            // If there's a body, parse it as JSON and check for "message" field.
            using TextReader reader = new StreamReader(request.Body);
            string text = await reader.ReadToEndAsync();

            _logger.LogInformation(text);
            if (text.Length > 0)
            {
                try
                {
                    JsonElement json = JsonSerializer.Deserialize<JsonElement>(text);
                    // if (json.TryGetProperty("links.dashboard_download_direct.href", out JsonElement messageElement) &&
                    //     messageElement.ValueKind == JsonValueKind.String)
                    // {
                    //     _logger.LogInformation("TryGet succeeded");
                    //     message = messageElement.GetString();
                    // }
                    if (json.TryGetProperty("buildNumber", out JsonElement messageElement))
                    {
                        _logger.LogInformation("Got build number");
                        if (messageElement.ValueKind == JsonValueKind.String)
                        {
                            string test = messageElement.GetString();
                            _logger.LogInformation($"and it's a string! {test}");
                        }
                    }
                }
                catch (JsonException parseException)
                {
                    _logger.LogError(parseException, "Error parsing JSON request");
                }
            }

            _logger.LogInformation($"Href: {message}");

            await context.Response.WriteAsync(message ?? "Hello World!");
        }
    }
}

