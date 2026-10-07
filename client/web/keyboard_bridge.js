// Flutter's raw-key response is asynchronous. Firefox can perform native
// Tab traversal before that response cancels it, overriding Flutter focus.
// Cancel synchronously only when the normal Flutter action handled Tab;
// unhandled traversal still leaves the view through the browser normally.
(() => {
  let tabEvent = null;
  window.addEventListener('keydown', event => {
    tabEvent = null;
    if (event.key !== 'Tab' || event.altKey || event.ctrlKey || event.metaKey ||
        !event.composedPath().some(node => node.tagName === 'FLUTTER-VIEW')) return;
    tabEvent = event;
    // Chromium may drain microtasks between native event listeners. Retain
    // this event for its whole dispatch, then release it at the next task.
    setTimeout(() => { if (tabEvent === event) tabEvent = null; }, 0);
  }, true);
  window.cgmsCompleteTabTraversal = handled => {
    // An asynchronous or unrelated action must never cancel a finished event.
    if (handled && tabEvent && tabEvent.eventPhase !== 0) tabEvent.preventDefault();
  };
})();
