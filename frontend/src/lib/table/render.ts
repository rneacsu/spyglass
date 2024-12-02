import { translateTableCell } from "$lib/translator";
import type { ConfigColumns } from "datatables.net-bs5"
import dayjs from "dayjs";


type RenderDecorator = (previous: string | HTMLElement, data: any, row: any) => string | HTMLElement
type RenderFilter = (data: any, row: any) => any
type RenderSort = (data: any, row: any) => any

export class RenderBuilder {

  private decorators: RenderDecorator[] = []
  private filterFn: RenderFilter = (data) => data
  private sortFn: RenderSort = (data) => data

  private constructor() {}

  static create() {
    return new RenderBuilder()
  }

  decorate(decorator: RenderDecorator) {
    this.decorators.push(decorator)
    return this
  }

  filter(filter: RenderFilter) {
    this.filterFn = filter
    return this
  }

  sort(sort: RenderSort) {
    this.sortFn = sort
    return this
  }

  build(): ConfigColumns['render'] {
    return (data: any, type: string, row: any) => {
      switch (type) {
        case 'sort':
        case 'type':
          return this.sortFn(data, row)
        case 'filter':
          return this.filterFn(data, row)
        case 'display':
          return this.decorators.reduce((previous, decorator) => decorator(previous, data, row), data)
        default:
          return data
      }
    }
  }
}

let ellipsisDecorator: RenderDecorator = (previous: string | HTMLElement) => {
    const span = document.createElement("span");
    span.classList.add("text-truncate");
    span.classList.add("d-inline-block");

    if (previous instanceof HTMLElement) {
      span.appendChild(previous);
      span.title = previous.title ?? previous.textContent;
    } else {
      span.textContent = previous;
      span.title = previous;
    }
    return span;
}


let translateDecorator: RenderDecorator = (previous: string | HTMLElement) => {
  return translateTableCell(previous instanceof HTMLElement ? (previous.textContent ?? "") : previous);
}

export function renderDefault() {
  return RenderBuilder.create()
    .decorate(translateDecorator)
    .decorate(ellipsisDecorator)
    .build()
}

export function renderRelativeTime() {
  return RenderBuilder.create()
    .decorate((_previous, data) => {
      const now = new Date();
      const past = new Date(Number(data) * 1000);
      let diff = now.getTime() - past.getTime();
      let future = false;

      if (diff < 0) {
        future = true;
        diff *= -1;
      }

      const seconds = Math.floor(diff / 1000);
      const minutes = Math.floor(seconds / 60);
      const hours = Math.floor(minutes / 60);
      const days = Math.floor(hours / 24);
      const years = Math.floor(days / 365);

      let result = "";

      if (years > 0) {
        result = `${years}y`;
      } else if (days > 0) {
        result = `${days}d`;
      } else if (hours > 0) {
        result = `${hours}h`;
      } else if (minutes > 0) {
        result = `${minutes}m`;
      } else {
        result = `${seconds}s`;
      }

      const relativeTime = future ? `in ${result}` : `${result} ago`;

      const span = document.createElement("span");
      span.title = dayjs(past).format();
      span.textContent = relativeTime;
      span.dataset.ageTimestamp = data;
      return span;

    })
    .decorate(ellipsisDecorator)
    .build()
}

export function renderStatus() {
  const statusMap = {
    "Completed": "info",
    "Succeeded": "info",
    "Running": "success",
    "Pending": "warning",
    "OOMKilled": "danger",
    "Failed": "danger",
    "CrashLoopBackOff": "danger",
  } as { [key: string]: string };


  return RenderBuilder.create()
    .decorate(translateDecorator)
    .decorate((previous, data) => {
      if (previous instanceof HTMLElement) {
        return previous
      }

      const span = document.createElement("span");
      span.classList.add("badge");
      span.classList.add("bg-" + (statusMap[data] ?? "secondary"));
      span.textContent = previous;
      span.title = data;
      return span;
    })
    .sort((data) => {
      const i = Object.keys(statusMap).indexOf(data);
      return i === -1 ? `999${data}` : i.toString().padStart(3, "0");
    })
    .build()
}

export function renderSelector() {
  return RenderBuilder.create()
    .decorate((previous) => {
      if (previous instanceof HTMLElement) {
        return previous;
      }

      const spanContainer = document.createElement("span");
      spanContainer.title = previous;

      previous.split(",").map((item) => {
        const parts = item.split("=");
        const key = parts[0];
        const value = parts[1];

        const span = document.createElement("span");
        span.classList.add("badge");
        span.classList.add("rounded-pill");
        span.classList.add("bg-primary");
        span.textContent = `${key}=${value}`;
        spanContainer.appendChild(span);
      })

      return spanContainer;
    })
    .decorate(ellipsisDecorator)
    .build()
}
