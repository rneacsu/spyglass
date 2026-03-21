import { createClient, type Client } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { Kube } from "./proto/kube_pb";

declare global {
  interface Window {
    GetGRPCUrl: () => Promise<string>;
  }
}

class GRPCClientWrapper {
  client: Client<typeof Kube>;

  constructor(baseUrl: string) {
    const transport = createConnectTransport({ baseUrl })
    this.client = createClient(Kube, transport)
  }
}

let wrapper: GRPCClientWrapper | null = null;

export default (async () => {
  if (!wrapper) {
    const url = await window.GetGRPCUrl();
    wrapper = new GRPCClientWrapper(url);
  }
  return wrapper.client;
})()
