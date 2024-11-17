import { Code, ConnectError } from "@connectrpc/connect";

export interface RefresherHandler {
  refresh(signal: AbortSignal): Promise<void>;
  onError(err: ConnectError): void;
}

export class Refresher {

  private abortController: AbortController | null = null;
  private refreshCancel: number | null = null;

  constructor(
    private handler: RefresherHandler,
    private interval: number = 5000,
  ) {

  }

  public abort() {
    if (this.abortController) {
      this.abortController.abort();
    }
    if (this.refreshCancel) {
      clearTimeout(this.refreshCancel);
      this.refreshCancel = null;
    }
  }

  private afterRefresh() {
    this.abortController = null;
    this.refreshCancel = setTimeout(() => {
      this.refresh();
    }, this.interval);
  }

  public async refresh() {
    this.abort();

    this.abortController = new AbortController();

    await this.handler.refresh(this.abortController.signal)
      .then(() => {
        this.afterRefresh();
      })
      .catch((err) => {
        const connErr = ConnectError.from(err)
        if (connErr.code === Code.Canceled) {
          return;
        }

        this.handler.onError(connErr);
        this.afterRefresh();
      })
  }
}
