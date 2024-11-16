<script lang="ts">
  import Alerts from "$lib/alerts.svelte";
  import Navbar from "$lib/navbar.svelte";
  import Sidebar from "$lib/sidebar.svelte";
  import Table from "$lib/table.svelte";

  let selectedContext: string = "";
  let selectedNamespace: string = "";

  let selectedGroup: string = "";
  let selectedVersion: string = "";
  let selectedResource: string = "";
  let selectedNamespaced: boolean = false;
</script>

<div class="h-100 d-flex flex-column">
  <Navbar
    bind:selectedContext
    bind:selectedNamespace
    namespaced={selectedNamespaced}
  />
  <Alerts />

  <div class="d-flex flex-row h-0 flex-grow-1">
    <Sidebar
      context={selectedContext}
      bind:selectedGroup
      bind:selectedVersion
      bind:selectedResource
      bind:selectedNamespaced
    />

    <div
      class="container-fluid py-2 flex-grow-1 w-0 position-relative overflow-y-hidden"
    >
      <div class="backlogo" />
      <Table
        context={selectedContext}
        namespace={selectedNamespace}
        group={selectedGroup}
        version={selectedVersion}
        resource={selectedResource}
        namespaced={selectedNamespaced}
      />
    </div>
  </div>
</div>

<style lang="scss">
  @import "bootswatch/dist/flatly/variables";

  .backlogo {
    position: absolute;
    left: 50%;
    top: 50%;
    width: 200px;
    height: 200px;
    background: $gray-800;
    transform: translate(-50%, -50%);
    z-index: -100;

    mask-image: url("$lib/assets/k8s.svg");
    mask-size: contain;
    mask-repeat: no-repeat;
    mask-position: center;
  }
</style>
