import {selection} from './selection.js';
const themeButton=document.querySelector('#theme');
function syncTheme(){const light=document.documentElement.dataset.theme==='light';themeButton.textContent=light?'☾':'☀';themeButton.setAttribute('aria-label',`Switch to ${light?'dark':'light'} mode`)}
themeButton.hidden=false;syncTheme();themeButton.addEventListener('click',()=>{const value=document.documentElement.dataset.theme==='light'?'dark':'light';document.documentElement.dataset.theme=value;try{localStorage.setItem('eitri-theme',value)}catch{}syncTheme()});
for(const button of document.querySelectorAll('.copy')){button.hidden=false;button.addEventListener('click',async()=>{const block=button.closest('.code-block'),status=block.querySelector('.copy-status');try{await navigator.clipboard.writeText(block.querySelector('code').textContent);status.textContent='Copied to clipboard.'}catch{status.textContent='Copy unavailable. Select the text above and copy it manually.'}})}
const play=document.querySelector('#play');if(play){play.hidden=false;play.addEventListener('click',()=>{const video=document.createElement('video');video.controls=true;video.playsInline=true;video.poster='/assets/workflow.jpg';video.setAttribute('aria-label','Eitri workflow recording');const track=document.createElement('track');track.kind='captions';track.label='English';track.srclang='en';track.src='/assets/captions.vtt';track.default=false;video.append(track);video.src='/assets/workflow.mp4';document.querySelector('#demo').replaceChildren(video);video.focus();video.play().catch(()=>{})})}
const language=document.querySelector('#lang');if(language){
 document.documentElement.classList.add('enhanced');
 for(const el of document.querySelectorAll('.guide-controls,.step-nav,.step-actions'))el.hidden=false;
 let step=0,all=false;
 const os=document.querySelector('#os'),variants=[...document.querySelectorAll('.variant')],go=[...document.querySelectorAll('[data-go]')],prev=document.querySelector('#previous'),next=document.querySelector('#next'),toggle=document.querySelector('#show-all');
 function render(focus=false){
  const value=selection(location.search);language.value=value.lang;os.value=value.os;
  for(const v of variants){const active=v.dataset.lang===value.lang&&v.dataset.os===value.os;v.hidden=!active;for(const s of v.querySelectorAll('.step'))s.hidden=!all&&Number(s.dataset.step)!==step}
  for(const b of go){if(Number(b.dataset.go)===step)b.setAttribute('aria-current','step');else b.removeAttribute('aria-current')}
  prev.disabled=step===0;next.disabled=step===6;document.querySelector('#step-count').textContent=all?'All 7 steps':`${step+1} / 7`;
  prev.hidden=next.hidden=all;toggle.textContent=all?'One step at a time':'Show all steps';toggle.setAttribute('aria-pressed',String(all));
  if(focus)document.querySelector('.variant:not([hidden]) .step:not([hidden]) h2').focus();
 }
 function choose(){const url=new URL(location.href);url.searchParams.set('lang',language.value);url.searchParams.set('os',os.value);history.pushState({},'',url);step=0;render()}
 language.addEventListener('change',choose);os.addEventListener('change',choose);window.addEventListener('popstate',()=>{step=0;render()});
 prev.addEventListener('click',()=>{step=Math.max(0,step-1);render(true)});next.addEventListener('click',()=>{step=Math.min(6,step+1);render(true)});
 for(const b of go)b.addEventListener('click',()=>{step=Number(b.dataset.go);render();const active=document.querySelector('.variant:not([hidden])');const target=active.querySelector(`[data-step="${step}"] h2`);target.focus();target.scrollIntoView({block:'start'})});
 toggle.addEventListener('click',()=>{all=!all;render()});render();
}
