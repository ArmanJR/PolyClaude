import SwiftUI
import os

@main
struct ClaudeUsagesApp: App {
    private static let logger = Logger(subsystem: "com.polyclaude.ClaudeUsages", category: "App")

    @State private var service: UsageService = {
        let s = UsageService()
        s.startPolling()
        return s
    }()

    var body: some Scene {
        MenuBarExtra {
            UsageMenuView(service: service)
        } label: {
            MenuBarLabel(service: service)
        }
        .menuBarExtraStyle(.window)

        Settings {
            SettingsView(service: service)
        }
    }
}

/// Menu bar icon label. Must be a real `View` (not a computed property on `App`)
/// so SwiftUI's `@Observable` tracking re-renders it when `service.accounts` or
/// `service.activeEmail` change during background polling.
private struct MenuBarLabel: View {
    let service: UsageService

    var body: some View {
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
