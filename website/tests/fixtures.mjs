import {execFileSync} from 'node:child_process';
import {mkdtempSync,writeFileSync,rmSync} from 'node:fs';
import {tmpdir} from 'node:os';
import {join} from 'node:path';
import {languages,compiler} from '../content.mjs';
import assert from 'node:assert/strict';
import {fileURLToPath} from 'node:url';
process.chdir(fileURLToPath(new URL('../../',import.meta.url)));
execFileSync('go',['run','website/tests/validate.go'],{stdio:'inherit'});
for(const [lang,l] of Object.entries(languages)){
 const dir=mkdtempSync(join(tmpdir(),'eitri-example-'));
 try {
  const c=compiler(lang,process.platform==='win32'?'windows':'macos');
  writeFileSync(join(dir,c.source),l.code);
  if(c.compile)execFileSync(c.compile,c.args,{cwd:dir,timeout:30000});
  const cmd=lang==='python'?(process.platform==='win32'?'python':'python3'):join(dir,process.platform==='win32'?'solution.exe':'solution');
  for(const [input,expected] of [['8\n','YES'],['2\n','NO'],['3\n','NO'],['4\n','YES']]){
   const actual=execFileSync(cmd,lang==='python'?[c.source]:[],{input,cwd:dir,timeout:3000,encoding:'utf8'}).trim();
   assert.equal(actual,expected,`${lang} input ${input}`);
  }
  console.log(`${l.name}: sample and three boundary cases pass (${process.platform}).`);
 }finally{rmSync(dir,{recursive:true,force:true})}
}
