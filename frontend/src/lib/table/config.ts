export type RenderType = "status" | "selector" | "timestamp";

type TableConfig = {
  hiddenColumns: string[],
  columnOrder: string[],
  showName: boolean,
  showAge: boolean,
  render: { [key: string]: RenderType },
  defaultOrder: { col: string, dir: "asc" | "desc" }[];
}

type TableOverride = {
  [Prop in keyof TableConfig]?: TableConfig[Prop]
}

export function mergeOverrides(...overrides: TableOverride[]): TableConfig {
  return overrides.reduce<TableConfig>((acc, override) => {
    return {
      hiddenColumns: [...(acc.hiddenColumns || []), ...(override.hiddenColumns || [])],
      columnOrder: [...(acc.columnOrder || []), ...(override.columnOrder || [])],
      render: { ...acc.render, ...override.render },
      showName: override.showName ?? acc.showName,
      showAge: override.showAge ?? acc.showAge,
      defaultOrder: override.defaultOrder ?? acc.defaultOrder
    };
  }, {
    hiddenColumns: [],
    columnOrder: [],
    render: {},
    showName: true,
    showAge: true,
    defaultOrder: []
  });
}

export function getConfig(group: string, version: string, resource: string): TableConfig {
  return mergeOverrides(overrides["*"], overrides[`${group}/${version}::${resource}`] ?? {});
}


let overrides: { [key: string]: TableOverride } = {
  "*": {
    hiddenColumns: ["Name", "Namespace", "Age"],
    showAge: true,
    showName: true,
    defaultOrder: [
      { col: "Name", dir: "asc" }
    ]
  },
  "/v1::pods": {
    hiddenColumns: ["Nominated Node", "Readiness Gates"],
    columnOrder: ["Status", "Ready"],
    render: {
      "Status": "status",
    }
  },
  "/v1::nodes": {
    hiddenColumns: ["Kernel-Version", "OS-Image", "Container-Runtime"],
  },
  "/v1::events": {
    hiddenColumns: ["Subobject"],
    columnOrder: ["Message", "Source", "Type", "Reason", "First Seen", "Last Seen"],
    showName: false,
    showAge: false
  },
  "apps/v1::deployments": {
    hiddenColumns: ["Images", "Containers"],
    render: {
      "Selector": "selector"
    }
  },
  "apps/v1::statefulsets": {
    hiddenColumns: ["Images", "Containers"],
  },
  "apps/v1::daemonsets": {
    hiddenColumns: ["Node Selector", "Containers", "Images"],
    render: {
      "Selector": "selector"
    }
  },
  "apps/v1::replicasets": {
    hiddenColumns: ["Images", "Containers"],
    render: {
      "Selector": "selector"
    }
  }
};
