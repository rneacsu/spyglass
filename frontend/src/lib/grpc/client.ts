import { createConnectTransport } from "@connectrpc/connect-web";
import { GetGRPCUrl } from "$lib/wailsjs/go/app/AppApi";
import { LogInfo } from "$lib/wailsjs/runtime/runtime";
import { createPromiseClient, type PromiseClient } from "@connectrpc/connect";
import { Greeter } from "./proto/greet_connect";

class GRPCClientWrapper {
  client: PromiseClient<typeof Greeter>;

  constructor(baseUrl: string) {
    const transport = createConnectTransport({ baseUrl })
    this.client = createPromiseClient(Greeter, transport)
  }
}

let wrapper: GRPCClientWrapper | null = null;

export default (async () => {
  if (!wrapper) {
    const url = await GetGRPCUrl();
    LogInfo("Creating GRPC client with url " + url);
    wrapper = new GRPCClientWrapper(url);
  }
  return wrapper.client;
})()
