{{flutter_js}}
{{flutter_build_config}}
const stage = document.getElementById('stage');
const retry = document.getElementById('retry');
retry.addEventListener('click', () => location.reload());
const slow = setTimeout(() => {
  stage.textContent = 'Taking longer than usual. You can retry loading.';
  retry.hidden = false;
}, 10000);
_flutter.loader.load({onEntrypointLoaded: async engine => {
  try {
    stage.textContent = 'Preparing the table…';
    const runner = await engine.initializeEngine();
    await runner.runApp();
    clearTimeout(slow);
    document.getElementById('loading')?.remove();
  } catch (_) {
    clearTimeout(slow);
    stage.textContent = 'The game could not start. Please retry loading.';
    retry.hidden = false;
  }
}});
