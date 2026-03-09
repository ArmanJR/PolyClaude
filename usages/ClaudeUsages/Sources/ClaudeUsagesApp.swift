import SwiftUI
import os

@main
struct ClaudeUsagesApp: App {
    private static let logger = Logger(subsystem: "com.polyclaude.ClaudeUsages", category: "App")

    @State private var service = UsageService()

    var body: some Scene {
        MenuBarExtra("Claude Usages", systemImage: "gauge.medium") {
            UsageMenuView(service: service)
                .task {
                    Self.logger.info("Menu opened — fetching usage data")
                    await service.fetch()
                }
        }
        .menuBarExtraStyle(.window)

        Settings {
            SettingsView(service: service)
        }
    }
}
