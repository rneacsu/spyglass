#include <string>
#include <semaphore>

#include <windows.h>
#include <wrl.h>

#include "WebView2.h"

static ICoreWebView2Environment *wv2_env;

using namespace Microsoft::WRL;
using namespace Microsoft::WRL::Wrappers;

extern "C" {

int wv2_init() {
  std::binary_semaphore sem{0};

  HRESULT hr = CreateCoreWebView2EnvironmentWithOptions(nullptr, nullptr, nullptr,
    Microsoft::WRL::Callback<ICoreWebView2CreateCoreWebView2EnvironmentCompletedHandler>(
      [&sem](HRESULT result, ICoreWebView2Environment *env) -> HRESULT {
        wv2_env = env;
        sem.release();
        return S_OK;
      }).Get());

  if (!SUCCEEDED(hr)) {
    return -1;
  }

  sem.acquire();

  return 0;
}

}
