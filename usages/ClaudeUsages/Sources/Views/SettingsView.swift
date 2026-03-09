import SwiftUI
import os

struct SettingsView: View {
    private static let logger = Logger(subsystem: "com.polyclaude.ClaudeUsages", category: "Settings")

    @AppStorage("serverURL") private var serverURL = "http://pi:8080/claude-usages"
    @AppStorage("fontSize") private var fontSize: Double = 14
    @State private var testResult: (success: Bool, message: String)?
    @State private var isTesting = false

    let service: UsageService

    var body: some View {
        Form {
            Section("Server") {
                TextField("URL", text: $serverURL)
                    .textFieldStyle(.roundedBorder)

                HStack {
                    Button("Test Connection") {
                        Task { await testConnection() }
                    }
                    .disabled(isTesting || serverURL.isEmpty)

                    if isTesting {
                        ProgressView()
                            .controlSize(.small)
                    }

                    if let result = testResult {
                        Image(systemName: result.success ? "checkmark.circle.fill" : "xmark.circle.fill")
                            .foregroundStyle(result.success ? .green : .red)
                        Text(result.message)
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }
            }

            Section("Appearance") {
                HStack {
                    Text("Font Size: \(Int(fontSize))pt")
                    Slider(value: $fontSize, in: 11...20, step: 1)
                }
            }
        }
        .formStyle(.grouped)
        .frame(width: 400, height: 220)
    }

    private func testConnection() async {
        Self.logger.info("Testing connection to \(serverURL)")
        isTesting = true
        testResult = nil
        let (success, message) = await service.testConnection(url: serverURL)
        testResult = (success, message)
        isTesting = false
        Self.logger.info("Connection test result: success=\(success), message=\(message)")
    }
}
