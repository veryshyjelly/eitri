import {readFileSync} from 'node:fs';
// Shared explanatory examples; also exercised by the Go verification harness.
export const transformations=JSON.parse(readFileSync(new URL('./tests/fixtures/transformations.json',import.meta.url),'utf8'));
