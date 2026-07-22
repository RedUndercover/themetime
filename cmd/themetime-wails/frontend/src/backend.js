export function backend() {
  return window.go?.main?.App;
}

export async function waitForBackend() {
  for (let attempt = 0; attempt < 80; attempt += 1) {
    if (backend()) return;
    await new Promise((resolve) => window.setTimeout(resolve, 50));
  }
  throw new Error('Wails bindings are not available.');
}
