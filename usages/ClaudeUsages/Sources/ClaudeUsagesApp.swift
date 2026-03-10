import SwiftUI
import os

@main
struct ClaudeUsagesApp: App {
    private static let logger = Logger(subsystem: "com.polyclaude.ClaudeUsages", category: "App")

    @State private var service = UsageService()

    var body: some Scene {
        MenuBarExtra {
            UsageMenuView(service: service)
                .task {
                    while !Task.isCancelled {
                        Self.logger.info("Polling — fetching usage data")
                        await service.fetch()
                        try? await Task.sleep(for: .seconds(60))
                    }
                }
        } label: {
            menuBarLabel
        }
        .menuBarExtraStyle(.window)

        Settings {
            SettingsView(service: service)
        }
    }

    @ViewBuilder
    private var menuBarLabel: some View {
        let parts = service.accounts.compactMap { account -> String? in
            guard let fiveHour = account.fiveHour else { return nil }
            let isActive = account.email == service.activeEmail
            let pct = "\(Int(fiveHour.effectiveUtilization))%"
            return isActive ? "[\(pct)]" : pct
        }
        if parts.isEmpty {
            Image(systemName: "gauge.medium")
        } else {
            Text(parts.joined(separator: " "))
                .font(.system(size: 11, weight: .medium, design: .monospaced))
        }
    }
}
