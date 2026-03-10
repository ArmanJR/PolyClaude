import SwiftUI

struct UsageMenuView: View {
    let service: UsageService
    @AppStorage("fontSize") private var fontSize: Double = 14
    @Environment(\.openSettings) private var openSettings

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            // Header
            HStack {
                Text("Claude Usages")
                    .font(.system(size: fontSize + 2, weight: .semibold))
                Spacer()
                Button {
                    NSApp.activate(ignoringOtherApps: true)
                    openSettings()
                } label: {
                    Image(systemName: "gear")
                        .font(.system(size: fontSize))
                }
                .buttonStyle(.plain)
                Button {
                    Task { await service.fetch() }
                } label: {
                    Image(systemName: "arrow.clockwise")
                        .font(.system(size: fontSize))
                }
                .buttonStyle(.plain)
                .disabled(service.isLoading)
                Button {
                    NSApplication.shared.terminate(nil)
                } label: {
                    Image(systemName: "xmark.circle")
                        .font(.system(size: fontSize))
                }
                .buttonStyle(.plain)
            }
            .padding(.horizontal, 12)
            .padding(.top, 10)
            .padding(.bottom, 8)

            Divider()

            // Content
            ZStack {
                if service.isLoading && service.accounts.isEmpty {
                    VStack(spacing: 8) {
                        ProgressView()
                            .controlSize(.small)
                        Text("Loading...")
                            .font(.system(size: fontSize - 1))
                            .foregroundStyle(.secondary)
                    }
                    .frame(maxWidth: .infinity, minHeight: 100)
                } else if let error = service.errorMessage, service.accounts.isEmpty {
                    VStack(spacing: 8) {
                        Image(systemName: "wifi.exclamationmark")
                            .font(.system(size: fontSize + 6))
                            .foregroundStyle(.secondary)
                        Text(error)
                            .font(.system(size: fontSize - 1))
                            .foregroundStyle(.secondary)
                            .multilineTextAlignment(.center)
                            .padding(.horizontal)
                    }
                    .frame(maxWidth: .infinity, minHeight: 100)
                } else {
                    VStack(alignment: .leading, spacing: 0) {
                        ForEach(Array(service.accounts.enumerated()), id: \.element.id) { index, account in
                            AccountUsageRow(account: account, fontSize: fontSize, isActive: account.email == service.activeEmail)
                                .padding(.horizontal, 12)
                            if index < service.accounts.count - 1 {
                                Divider()
                                    .padding(.horizontal, 12)
                            }
                        }
                    }
                    .padding(.vertical, 6)
                    .overlay {
                        // Overlay spinner when refreshing with existing data
                        if service.isLoading {
                            VStack {
                                HStack {
                                    Spacer()
                                    ProgressView()
                                        .controlSize(.mini)
                                        .padding(6)
                                }
                                Spacer()
                            }
                        }
                    }
                    .overlay(alignment: .bottom) {
                        // Show error banner if we have data but also an error
                        if let error = service.errorMessage {
                            HStack {
                                Image(systemName: "exclamationmark.triangle")
                                    .font(.system(size: fontSize - 2))
                                Text(error)
                                    .font(.system(size: fontSize - 3))
                                    .lineLimit(1)
                            }
                            .foregroundStyle(.orange)
                            .padding(.horizontal, 12)
                            .padding(.vertical, 4)
                            .frame(maxWidth: .infinity, alignment: .leading)
                            .background(.ultraThinMaterial)
                        }
                    }
                }
            }
        }
        .frame(width: max(340, fontSize * 26))
    }


}
