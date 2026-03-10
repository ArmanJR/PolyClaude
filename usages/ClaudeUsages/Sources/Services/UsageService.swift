import Foundation
import os

@Observable
final class UsageService {
    private static let logger = Logger(subsystem: "com.polyclaude.ClaudeUsages", category: "UsageService")

    var accounts: [AccountUsage] = []
    var isLoading = false
    var errorMessage: String?
    var activeEmail: String?

    @ObservationIgnored
    private var serverURL: String = UserDefaults.standard.string(forKey: "serverURL")
        ?? "http://pi:8080/claude-usages"

    func fetch() async {
        let url = UserDefaults.standard.string(forKey: "serverURL") ?? "http://pi:8080/claude-usages"
        serverURL = url
        Self.logger.info("Fetching usage data from \(url)")

        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        guard let requestURL = URL(string: url) else {
            Self.logger.error("Invalid server URL: \(url)")
            errorMessage = "Invalid server URL"
            return
        }

        do {
            let (data, response) = try await URLSession.shared.data(from: requestURL)
            guard let httpResponse = response as? HTTPURLResponse else {
                Self.logger.error("Non-HTTP response received")
                errorMessage = "Invalid response from server"
                return
            }

            guard httpResponse.statusCode == 200 else {
                Self.logger.error("Server returned HTTP \(httpResponse.statusCode)")
                errorMessage = "Server error: HTTP \(httpResponse.statusCode)"
                return
            }

            let decoder = JSONDecoder()
            let dict = try decoder.decode([String: AccountUsage].self, from: data)

            var result: [AccountUsage] = []
            for (email, var usage) in dict {
                usage.email = email
                result.append(usage)
            }
            result.sort {
                ($0.fiveHour?.utilization ?? -1) < ($1.fiveHour?.utilization ?? -1)
            }

            accounts = result
            activeEmail = Self.readActiveEmail()
            Self.logger.info("Fetched \(result.count) account(s), active: \(self.activeEmail ?? "unknown")")
        } catch let error as URLError {
            Self.logger.error("Network error: \(error.localizedDescription)")
            errorMessage = "Cannot reach server: \(error.localizedDescription)"
        } catch {
            Self.logger.error("Decode error: \(error.localizedDescription)")
            errorMessage = "Failed to parse response: \(error.localizedDescription)"
        }
    }

    private static func readActiveEmail() -> String? {
        let home = FileManager.default.homeDirectoryForCurrentUser
        let claudeJSON = home.appendingPathComponent(".claude.json")
        guard let data = try? Data(contentsOf: claudeJSON),
              let json = try? JSONSerialization.jsonObject(with: data) as? [String: Any],
              let oauth = json["oauthAccount"] as? [String: Any],
              let email = oauth["emailAddress"] as? String else {
            Self.logger.warning("Could not read active account from ~/.claude.json")
            return nil
        }
        return email
    }

    func testConnection(url: String) async -> (Bool, String) {
        Self.logger.info("Testing connection to \(url)")
        guard let requestURL = URL(string: url) else {
            return (false, "Invalid URL")
        }

        do {
            let (data, response) = try await URLSession.shared.data(from: requestURL)
            guard let httpResponse = response as? HTTPURLResponse else {
                return (false, "Invalid response")
            }
            guard httpResponse.statusCode == 200 else {
                return (false, "HTTP \(httpResponse.statusCode)")
            }
            let _ = try JSONDecoder().decode([String: AccountUsage].self, from: data)
            return (true, "Connected — valid response")
        } catch let error as URLError {
            return (false, error.localizedDescription)
        } catch {
            return (false, "Invalid JSON: \(error.localizedDescription)")
        }
    }
}
