<script lang="ts">
  import client from "./grpc/client";
  import { ConnectError } from "@connectrpc/connect";
  import Dropdown, { type Item } from "./dropdown.svelte";
  import { onDestroy, onMount, untrack } from "svelte";
  import { ShowAlert } from "./alerts.svelte";
  import { Refresher } from "./grpc/refresher";
  import { autoTheme } from "./autotheme/autoTheme.svelte";

  let {
    title = "SpyGlass",
    namespaced = false,
    context = $bindable(""),
    namespace = $bindable(""),
  } = $props();

  let contextItems: Item[] = $state([]);
  let namespaceItems: Item[] = $state([]);
  let namespaceLoading: boolean = $state(false);

  let namespaceRefresher: Refresher | null = null;

  $effect(() => {
    context !== null &&
      untrack(() => {
        onContextChange();
      });
  });

  async function loadNamespaces(abort: AbortSignal) {
    if (!context) {
      return;
    }

    // ShowAlert("info", "Loading namespaces...");

    const resources = (
      await (
        await client
      ).listResource(
        {
          context: context,
          gvr: {
            group: "",
            version: "v1",
            resource: "namespaces",
          },
        },
        { signal: abort },
      )
    ).resources;

    if (resources.length > 0) {
      namespaceItems = resources.map((ns) => ({
        label: ns.name,
        value: ns.name,
      }));
      namespaceItems.unshift({ label: "All", value: "__all__" });
    }
  }

  function onContextChange() {
    namespaceItems = [];

    (async () => {
      namespaceLoading = true;
      await namespaceRefresher?.refresh();
      namespaceLoading = false;
    })();
  }

  onMount(() => {
    client
      .then(async (client) => {
        // Make sure to await all promises before setting the state
        const contextsResponse = await client.getContexts({});
        const defaultContextResponse = await client.getDefaultContext({});

        contextItems = contextsResponse.contexts.map((context) => ({
          label: context,
          value: context,
        }));
        context = defaultContextResponse.context;

        namespaceRefresher = new Refresher({
          refresh: loadNamespaces,
          onError: (e) => {
            ShowAlert("error", e.message);
          },
        });
      })
      .catch((e) => {
        ShowAlert("error", ConnectError.from(e).message);
      });
  });

  onDestroy(() => {
    namespaceRefresher?.abort();
  });
</script>

<div>
  <nav class="navbar navbar-expand bg-primary" data-bs-theme="dark">
    <div class="container-fluid">
      <a
        class="navbar-brand"
        href={"#"}
        onclick={() => {
          ShowAlert("info", "This is a test alert");
        }}>{title}</a
      >
      <button
        class="navbar-toggler"
        type="button"
        data-bs-toggle="collapse"
        data-bs-target="#navbarSupportedContent"
        aria-controls="navbarSupportedContent"
        aria-expanded="false"
        aria-label="Toggle navigation"
      >
        <span class="navbar-toggler-icon"></span>
      </button>
      <div
        class="collapse navbar-collapse"
        id="navbarSupportedContent"
        data-bs-theme={autoTheme.theme}
      >
        <ul class="navbar-nav me-auto mb-2 mb-lg-0"></ul>
        <div class="d-flex">
          {#if namespaced}
            <div class="me-3">
              <Dropdown
                alignEnd={true}
                isLoading={namespaceLoading}
                items={namespaceItems}
                bind:selectedItem={namespace}
                noItemsMessage="No namespaces"
                loadingMessage="Loading namespaces..."
                noSelectionMessage="Select namespace..."
              />
            </div>
          {/if}
          <Dropdown
            alignEnd={true}
            items={contextItems}
            bind:selectedItem={context}
            noItemsMessage="No contexts"
            loadingMessage="Loading contexts..."
            noSelectionMessage="Select context..."
          />
        </div>
      </div>
    </div>
  </nav>
</div>

<style>
</style>
