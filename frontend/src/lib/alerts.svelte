<script lang="ts" module>
  export class Alert {
    id: number;
    message: string;
    type: AlertType;
    timeout: number;
    timeoutId: number | null = null;

    static autoId = 0;

    constructor(message: string, type: AlertType = "default", timeout = 5000) {
      this.message = message;
      this.type = type;
      this.id = Alert.autoId++;
      this.timeout = timeout;
    }

    getAlertClass(): string {
      switch (this.type) {
        case "default":
          return "primary";
        case "success":
          return "success";
        case "info":
          return "info";
        case "warning":
          return "warning";
        case "error":
          return "danger";
        default:
          return "secondary";
      }
    }
  }
  export type AlertType = "default" | "success" | "info" | "warning" | "error";

  let alerts: Alert[] = $state([]);

  function addAlert(alert: Alert) {
    alerts.unshift(alert);

    if (alert.timeout > 0) {
      alert.timeoutId = window.setTimeout(() => {
        removeAlert(alert.id);
      }, alert.timeout);
    }
  }

  function removeAlert(id: number) {
    const alert = alerts.find((a) => a.id === id);
    if (alert && alert.timeoutId) {
      clearTimeout(alert.timeoutId);
    }
    alerts.splice(
      alerts.findIndex((a) => a.id === id),
      1,
    );
  }

  export function ShowAlert(type: AlertType, message: string, timeout = 5000) {
    addAlert(new Alert(message, type, timeout));
  }
</script>

<script lang="ts">
  import { fade } from "svelte/transition";
</script>

<div class="alertsContainer d-flex p-3 flex-column-reverse align-items-end">
  {#each alerts as alert (alert.id)}
    <div
      class="alert alert-{alert.getAlertClass()} alert-dismissible mb-0 mt-3"
      role="alert"
      transition:fade
    >
      <div class="alertMessage">
        {alert.message}
      </div>
      <button
        type="button"
        class="btn-close"
        aria-label="Close"
        onclick={() => {
          removeAlert(alert.id);
        }}
      ></button>
    </div>
  {/each}
</div>

<style lang="scss">
  $maxHeight: 350px;
  $fadeLength: 100px;

  .alertsContainer {
    position: fixed;
    bottom: 0;
    right: 0;
    z-index: 1000;
    mask-image: linear-gradient(
      to top,
      rgba(0, 0, 0, 1) ($maxHeight - $fadeLength),
      rgba(0, 0, 0, 0) $maxHeight
    );
    max-height: $maxHeight;
    overflow-y: hidden;
  }
  .alertMessage {
    display: -webkit-inline-box;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 4;
    max-width: 650px;
    line-clamp: 4;
    text-overflow: ellipsis;
    overflow: hidden;
  }
</style>
