package sim

import (
	"os/exec"
	"path/filepath"
	"testing"
)

func TestEvaluationHelperRejectsCorruption(t *testing.T) {
	root, e := filepath.Abs("../../..")
	if e != nil {
		t.Fatal(e)
	}
	script := `import importlib.util,json,tempfile,pathlib,hashlib
spec=importlib.util.spec_from_file_location('evaluation','xops/sim_evaluate.py');m=importlib.util.module_from_spec(spec);spec.loader.exec_module(m)
with tempfile.TemporaryDirectory() as tmp:
 p=pathlib.Path(tmp); data=b'Observer: public. Finalized games: 0 / 1.';(p/'report.md').write_bytes(data)
 manifest={'schema_version':'cgms-manifest-v1','exit_code':3,'status':'unsupported','privacy':'restricted','root_seed':'private-canary','artifacts':[{'path':'report.md','bytes':len(data),'sha256':hashlib.sha256(data).hexdigest(),'privacy':'public'}]}
 (p/'manifest.json').write_text(json.dumps(manifest));m.inspect_manifest(p,3)
 for mode in ('digest','exit','privacy','escape'):
  x=json.loads(json.dumps(manifest))
  if mode=='digest': x['artifacts'][0]['sha256']='0'*64
  if mode=='exit': x['exit_code']=0
  if mode=='privacy': (p/'report.md').write_text('private-canary');x['artifacts'][0]['bytes']=14;x['artifacts'][0]['sha256']=hashlib.sha256(b'private-canary').hexdigest()
  if mode=='escape': x['artifacts'][0]['path']='../outside'
  (p/'manifest.json').write_text(json.dumps(x))
  try: m.inspect_manifest(p,3)
  except (ValueError,AssertionError,FileNotFoundError): pass
  else: raise AssertionError('corrupt '+mode+' accepted')
`
	cmd := exec.Command("python3", "-c", script)
	cmd.Dir = root
	if b, e := cmd.CombinedOutput(); e != nil {
		t.Fatalf("evaluation checker: %v\n%s", e, b)
	}
}
