"""Exercise real HTTP intake and save-triggered judging in a disposable PTY."""
import fcntl, json, os, pty, select, struct, subprocess, tempfile, termios, time, urllib.request
from pathlib import Path
repo = Path(__file__).resolve().parents[2]
fixtures = json.loads((repo/'website/dist/fixtures.json').read_text())
subprocess.run(['go','build','-o','/tmp/eitri-website-smoke','.'],cwd=repo,check=True)
with tempfile.TemporaryDirectory(prefix='eitri-smoke-') as d:
    root=Path(d); f=next(x for x in fixtures if x['lang']=='python' and x['os']=='macos')
    (root/'solutions').mkdir(); (root/'base.py').write_text(''); (root/'eitri.yml').write_text(f['config'])
    master,slave=pty.openpty(); fcntl.ioctl(slave,termios.TIOCSWINSZ,struct.pack('HHHH',35,110,0,0))
    child=subprocess.Popen(['/tmp/eitri-website-smoke'],cwd=root,stdin=slave,stdout=slave,stderr=slave,env={**os.environ,'TERM':'xterm-256color'})
    os.close(slave); captured=b''
    def drain():
        global captured
        while select.select([master],[],[],0)[0]:
            try: captured+=os.read(master,65536)
            except OSError: break
    try:
        # Wait for startup before sending the same shape as Competitive Companion.
        for _ in range(50):
            drain(); time.sleep(.1)
            try:
                data={'name':'A. Watermelon','url':f['problem']['url'],'memoryLimit':64,'timeLimit':1000,'tests':[{'input':'8\n','output':'YES\n'}]}
                req=urllib.request.Request('http://127.0.0.1:6174/',data=json.dumps(data).encode(),headers={'Content-Type':'application/json'})
                assert urllib.request.urlopen(req,timeout=1).status==200
                break
            except OSError:
                if child.poll() is not None: raise RuntimeError('Eitri exited before intake')
        else: raise RuntimeError('HTTP startup timed out')
        path=root/'solutions/4-A.py'
        for _ in range(30):
            drain();time.sleep(.1)
            if path.exists(): break
        assert path.exists(),'Source not created'
        time.sleep(.3); path.write_text(f['code'])
        for _ in range(80):
            drain();time.sleep(.1)
            if b'AC' in captured: break
        assert b'AC' in captured,'Save-triggered AC was not displayed'
        print('PASS: real HTTP intake, filename routing, file generation, save watcher, compilation/run, and AC rendering (macOS/Python).')
        print('Not exercised: browser extension click, judge submission, desktop clipboard.')
    finally:
        child.terminate()
        try:child.wait(timeout=3)
        except subprocess.TimeoutExpired:child.kill();child.wait()
        os.close(master)
