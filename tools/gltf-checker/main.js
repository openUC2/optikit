import { buffer } from "node:stream/consumers";
import { default as validator } from "gltf-validator";

const asset = await buffer(process.stdin);
let report;
try {
	report = await validator.validateBytes(new Uint8Array(asset));
} catch (error) {
	console.error(`Validation failed:`, error);
	process.exit(1);
}

const reportMetadata = { ...report };
reportMetadata.issues = { ...report.issues };
delete reportMetadata.issues.messages;
if (report.issues.numErrors !== 0) {
	console.info(`Asset failed with errors:`, reportMetadata);
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
	console.info(`Asset passed with warnings:`, reportMetadata);
	for (const message of report.issues.messages) {
		if (message.severity === 1) {
			console.warn(
				`  ${message.code} (${message.message}): ${message.pointer}`,
			);
		}
	}
	process.exit(0);
}

console.info(`Asset passed with no errors or warnings:`, reportMetadata);
for (const message of report.issues.messages) {
	switch (message.severity) {
		case 2:
			console.info(
				`  ${message.code} (${message.message}): ${message.pointer}`,
			);
			break;
		case 3:
			console.debug(
				`  ${message.code} (${message.message}): ${message.pointer}`,
			);
			break;
	}
}

process.exit(0);
