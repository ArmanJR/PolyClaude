import SwiftUI

struct AccountUsageRow: View {
    let account: AccountUsage
    let fontSize: Double
    var isActive: Bool = false

    private var labelFont: Font { .system(size: fontSize - 1) }
    private var valueFont: Font { .system(size: fontSize - 1, design: .monospaced) }
    private var subFont: Font { .system(size: fontSize - 3) }
    private var labelWidth: CGFloat { fontSize * 4.2 }
    private var percentWidth: CGFloat { fontSize * 3 }

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            // Header: email + error badge
            HStack {
                Text(account.email)
                    .font(.system(size: fontSize, weight: .semibold))
                    .lineLimit(1)
                if isActive {
                    Text("ACTIVE")
                        .font(.system(size: fontSize - 4, weight: .bold, design: .rounded))
                        .foregroundStyle(.white)
                        .padding(.horizontal, 5)
                        .padding(.vertical, 1)
                        .background(Color(red: 0.18, green: 0.55, blue: 0.24), in: Capsule())
                }
                if account.hasRecentError {
                    Image(systemName: "exclamationmark.triangle.fill")
                        .foregroundStyle(.yellow)
                        .font(.system(size: fontSize - 2))
                        .help("Data may be stale — server reported an error")
                }
            }

            Text(account.dirName)
                .font(.system(size: fontSize - 2))
                .foregroundStyle(.secondary)

            // 5-Hour limit
            if let fiveHour = account.fiveHour {
                usageLimitView(label: "5-Hour", limit: fiveHour)
            }

            // 7-Day limit
            if let sevenDay = account.sevenDay {
                usageLimitView(label: "7-Day", limit: sevenDay)
            }

            // Extra usage
            if let extra = account.extraUsage, extra.isEnabled {
                HStack {
                    Text("Extra")
                        .font(labelFont)
                        .foregroundStyle(.secondary)
                        .frame(width: labelWidth, alignment: .leading)
                    Text("\(extra.usedDollars) / \(extra.limitDollars)")
                        .font(valueFont)
                }
            }
        }
        .padding(.vertical, 4)
    }

    @ViewBuilder
    private func usageLimitView(label: String, limit: UsageLimit) -> some View {
        VStack(alignment: .leading, spacing: 2) {
            HStack(spacing: 6) {
                Text(label)
                    .font(labelFont)
                    .foregroundStyle(.secondary)
                    .frame(width: labelWidth, alignment: .leading)
                ProgressView(value: min(limit.utilization, 100), total: 100)
                    .tint(limit.utilizationColor)
                Text("\(Int(limit.utilization))%")
                    .font(valueFont)
                    .frame(width: percentWidth, alignment: .trailing)
            }
            HStack {
                Spacer().frame(width: labelWidth + 6)
                Text("Resets \(limit.resetsInFormatted)")
                    .font(subFont)
                    .foregroundStyle(.tertiary)
            }
        }
    }
}
