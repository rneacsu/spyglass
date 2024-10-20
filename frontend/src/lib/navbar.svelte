<script lang="ts">
  import client from "./grpc/client";
  import { ConnectError, Code } from "@connectrpc/connect";
  import Dropdown, { type Item } from "./dropdown.svelte";
  import { onDestroy, onMount } from "svelte";
  import { ShowAlert } from "./alerts.svelte";

  export let title = "SpyGlass";

  let contexts: string[] = [];
  let contextItems: Item[] = [];
  let selectedContext: string = "";

  let isLoadingNamespaces: boolean = false;

  let namespaces: string[] = [];
  let namespaceItems: Item[] = [];
  let selectedNamespace: string = "";

  let namespaceAbort: AbortController | null = null;
  let namespaceRefreshCancel: number | null = null;

  $: selectedContext && onContextChange();

  function loadNamespaces() {
    if (!selectedContext) {
      return;
    }
    if (namespaceAbort) {
      namespaceAbort.abort();
    }
    namespaceAbort = new AbortController();
    // ShowAlert("info", "Refreshing namespaces...");

    const currentRefreshCancel = namespaceRefreshCancel;

    client
      .then(async (client) => {
        const resources = (
          await client.listResource({
          context: selectedContext,
          gvr: {
            group: "",
            version: "v1",
            resource: "namespaces",
          }
        }, { signal: namespaceAbort?.signal })).resources;

        namespaces = resources.map((resource) => resource.name);

        if (namespaces.length > 0) {
          namespaceItems = namespaces.map((namespace) => ({
            label: namespace,
            value: namespace,
          }));
          namespaceItems.unshift({ label: "All", value: "__all__" });
        }
        isLoadingNamespaces = false;
        namespaceAbort = null;
      })
      .catch((e) => {
        const err = ConnectError.from(e)
        if (err.code != Code.Canceled) {
          ShowAlert("error", ConnectError.from(e).message);
          isLoadingNamespaces = false;
        }
      }).finally(() => {
        if (currentRefreshCancel !== namespaceRefreshCancel) {
          return;
        }
        namespaceRefreshCancel = window.setTimeout(() => {
          loadNamespaces();
        }, 5000);
      });
  }

  function onContextChange() {
    namespaces = [];
    namespaceItems = [];
    isLoadingNamespaces = true;
    if (namespaceRefreshCancel) {
      window.clearTimeout(namespaceRefreshCancel);
      namespaceRefreshCancel = null;
    }

    loadNamespaces();
  }

  onMount(() => {
    client
      .then(async (client) => {
        // Make sure to await all promises before setting the state
        const contextsResponse = await client.getContexts({});
        const defaultContextResponse = await client.getDefaultContext({});

        contexts = contextsResponse.contexts;
        contextItems = contexts.map((context) => ({
          label: context,
          value: context,
        }));
        selectedContext = defaultContextResponse.context;
      })
      .catch((e) => {
        ShowAlert("error", ConnectError.from(e).message);
      });
  });

  onDestroy(() => {
    if (namespaceAbort) {
      namespaceAbort.abort();
    }
    if (namespaceRefreshCancel) {
      window.clearTimeout(namespaceRefreshCancel);
    }
  });
</script>

<div>
  <nav class="navbar navbar-expand bg-primary">
    <div class="container-fluid">
      <a class="navbar-brand" href={"#"}
        on:click={() => {
          ShowAlert("info", "This is a test alert");
        }}
      >{title}</a>
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
      <div class="collapse navbar-collapse" id="navbarSupportedContent">
        <ul class="navbar-nav me-auto mb-2 mb-lg-0"></ul>
        <div class="d-flex">
          <div class="me-3">
            <Dropdown
              alignEnd={true}
              isLoading={isLoadingNamespaces}
              items={namespaceItems}
              bind:selectedItem={selectedNamespace}
              noItemsMessage="No namespaces"
              loadingMessage="Loading namespaces..."
              noSelectionMessage="Select namespace..."
            />
          </div>
          <Dropdown
            alignEnd={true}
            isLoading={contexts.length === 0}
            items={contextItems}
            bind:selectedItem={selectedContext}
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
