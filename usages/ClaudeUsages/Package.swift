// swift-tools-version: 5.10

import PackageDescription

let package = Package(
    name: "ClaudeUsages",
    platforms: [.macOS(.v14)],
    targets: [
        .executableTarget(
            name: "ClaudeUsages",
            path: "Sources"
        ),
    ]
)
