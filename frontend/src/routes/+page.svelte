<script lang="ts">
  import Alerts from "$lib/alerts.svelte";
  import Navbar from "$lib/navbar.svelte";
  import Sidebar from "$lib/sidebar/sidebar.svelte";
  import Table from "$lib/table.svelte";

  let selected = $state({
    context: "",
    namespace: "",
    group: "",
    version: "",
    resource: "",
    namespaced: false,
  });
</script>

<div class="h-100 d-flex flex-column">
  <Navbar
    bind:context={selected.context}
    bind:namespace={selected.namespace}
    namespaced={selected.namespaced}
  />
  <Alerts />

  <div class="d-flex flex-row h-0 flex-grow-1">
    <Sidebar
      context={selected.context}
      bind:group={selected.group}
      bind:version={selected.version}
      bind:resource={selected.resource}
      bind:namespaced={selected.namespaced}
    />

    <div
      class="container-fluid py-2 flex-grow-1 w-0 position-relative overflow-y-hidden"
    >
      <div class="backlogo"></div>
      <Table
        context={selected.context}
        namespace={selected.namespace}
        group={selected.group}
        version={selected.version}
        resource={selected.resource}
        namespaced={selected.namespaced}
      />
    </div>
  </div>
</div>

<style lang="scss">
  .backlogo {
    position: absolute;
    left: 50%;
    top: 50%;
    width: 200px;
    height: 200px;
    background: var(--bs-secondary-bg);
    transform: translate(-50%, -50%);
    z-index: -100;

    mask-image: url("$lib/assets/k8s.svg");
    mask-size: contain;
    mask-repeat: no-repeat;
    mask-position: center;
  }
</style>
