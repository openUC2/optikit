import { readFileSync } from "node:fs";
import { buffer } from "node:stream/consumers";
import { default as validator } from "gltf-validator";

const verbose = process.env.VERBOSE;
const input = process.env.INPUT;

let asset;
if (input) {
	asset = readFileSync(input);
} else {
	asset = await buffer(process.stdin);
}
const assetName = input ? input : "Asset";

let report;
try {
	report = await validator.validateBytes(new Uint8Array(asset));
} catch (error) {
	console.error(`${assetName} validation failed:`, error);
	process.exit(1);
}

const reportMetadata = { ...report };
reportMetadata.issues = { ...report.issues };
delete reportMetadata.issues.messages;
if (report.issues.numErrors !== 0) {
	console.info(`${assetName} failed with errors:`, reportMetadata);
	for (const message of report.issues.messages) {
		if (message.severity === 0) {
			console.error(
				`  ${message.code} (${message.message}): ${message.pointer}`,
			);
		}
	}
	process.exit(2);
}

if (report.issues.numWarnings !== 0) {
	console.info(`${assetName} passed with warnings:`, reportMetadata);
	for (const message of report.issues.messages) {
		if (message.severity === 1) {
			console.warn(
				`  ${message.code} (${message.message}): ${message.pointer}`,
			);
		}
	}
	process.exit(3);
}

if (report.issues.numInfos + report.issues.numHints > 0) {
	if (verbose) {
		console.info(
			`${assetName} passed with no errors or warnings:`,
			reportMetadata,
		);
	}
	for (const message of report.issues.messages) {
		switch (message.severity) {
			case 2:
				if (verbose) {
					console.info(
						`  ${message.code} (${message.message}): ${message.pointer}`,
					);
				}
				break;
			case 3:
				if (verbose) {
					console.debug(
						`  ${message.code} (${message.message}): ${message.pointer}`,
					);
				}
				break;
		}
	}
}

process.exit(0);
