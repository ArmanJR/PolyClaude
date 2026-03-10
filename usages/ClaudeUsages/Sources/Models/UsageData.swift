import Foundation
import SwiftUI

// MARK: - API Response Models

struct UsageLimit: Codable {
    let utilization: Double
    let resetsAt: String?

    enum CodingKeys: String, CodingKey {
        case utilization
        case resetsAt = "resets_at"
    }

    var resetsAtDate: Date? {
        guard let resetsAt else { return nil }
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        if let date = formatter.date(from: resetsAt) { return date }
        formatter.formatOptions = [.withInternetDateTime]
        return formatter.date(from: resetsAt)
    }

    var resetsInFormatted: String {
        guard let date = resetsAtDate else { return "unknown" }
        let now = Date()
        if date <= now { return "now" }

        let diff = Calendar.current.dateComponents([.day, .hour, .minute], from: now, to: date)
        if let days = diff.day, days > 0 {
            let formatter = DateFormatter()
            formatter.dateFormat = "MMM d"
            return formatter.string(from: date)
        }
        if let hours = diff.hour, hours > 0 {
            let mins = diff.minute ?? 0
            return "in \(hours)h \(mins)m"
        }
        if let mins = diff.minute, mins > 0 {
            return "in \(mins)m"
        }
        return "in <1m"
    }

    var isIdle: Bool {
        guard let date = resetsAtDate else { return true }
        return date <= Date()
    }

    var effectiveUtilization: Double {
        isIdle ? 0 : utilization
    }

    var utilizationColor: Color {
        switch utilization {
        case 0..<50: return .green
        case 50..<80: return .yellow
        case 80..<95: return .orange
        default: return .red
        }
    }
}

struct ExtraUsage: Codable {
    let isEnabled: Bool
    let monthlyLimit: Double?
    let usedCredits: Double?
    let utilization: Double?

    enum CodingKeys: String, CodingKey {
        case isEnabled = "is_enabled"
        case monthlyLimit = "monthly_limit"
        case usedCredits = "used_credits"
        case utilization
    }

    var usedDollars: String {
        let cents = usedCredits ?? 0.0
        return String(format: "$%.2f", cents / 100.0)
    }

    var limitDollars: String {
        let cents = monthlyLimit ?? 0.0
        return String(format: "$%.2f", cents / 100.0)
    }
}

struct AccountUsage: Codable, Identifiable {
    let dirName: String
    let fiveHour: UsageLimit?
    let sevenDay: UsageLimit?
    let extraUsage: ExtraUsage?
    let lastPullDatetime: String?
    let lastError: String?

    var email: String = ""
    var id: String { email }

    enum CodingKeys: String, CodingKey {
        case dirName = "dir_name"
        case fiveHour = "five_hour"
        case sevenDay = "seven_day"
        case extraUsage = "extra_usage"
        case lastPullDatetime = "last_pull_datetime"
        case lastError = "last_error"
    }

    var hasRecentError: Bool {
        guard let errorStr = lastError, let pullStr = lastPullDatetime else {
            return lastError != nil
        }
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        let errorDate = formatter.date(from: errorStr) ?? Date.distantPast
        let pullDate = formatter.date(from: pullStr) ?? Date.distantPast
        return errorDate > pullDate
    }

    var lastPullDate: Date? {
        guard let str = lastPullDatetime else { return nil }
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        if let date = formatter.date(from: str) { return date }
        formatter.formatOptions = [.withInternetDateTime]
        return formatter.date(from: str)
    }

    var lastFetchFormatted: String {
        guard let date = lastPullDate else { return "never" }
        let elapsed = Date().timeIntervalSince(date)
        if elapsed < 60 { return "just now" }
        if elapsed < 3600 {
            let mins = Int(elapsed / 60)
            return "\(mins)m ago"
        }
        let hours = Int(elapsed / 3600)
        return "\(hours)h ago"
    }
}
