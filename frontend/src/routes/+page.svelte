<script lang="ts">
  import Alerts from "$lib/alerts.svelte";
  import Navbar from "$lib/navbar.svelte";
</script>

<div class="h-100 d-flex flex-column">
  <Navbar />
  <Alerts />

  <div
    class="container-fluid py-2 flex-grow-1 position-relative overflow-y-hidden"
  >
    <div class="backlogo" />
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
