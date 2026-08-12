import { buffer } from "node:stream/consumers";
import { default as validator } from "gltf-validator";

const asset = await buffer(process.stdin);
try {
	const report = await validator.validateBytes(new Uint8Array(asset));
	console.info("Validation succeeded:", report);
} catch (error) {
	console.error("Validation failed:", error);
	process.exit(1);
}
