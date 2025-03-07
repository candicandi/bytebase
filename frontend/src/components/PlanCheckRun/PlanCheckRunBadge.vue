<template>
  <button
    class="inline-flex items-center px-3 leading-6 rounded-full text-sm border"
    :class="buttonClasses"
    @click="clickable && $emit('click')"
  >
    <template v-if="status === PlanCheckRun_Status.RUNNING">
      <TaskSpinner class="-ml-1 mr-1.5 h-4 w-4 text-info" />
    </template>
    <template v-else-if="status === PlanCheckRun_Status.DONE">
      <template v-if="resultStatus === PlanCheckRun_Result_Status.SUCCESS">
        <heroicons-outline:check
          class="-ml-1 mr-1.5 mt-0.5 h-4 w-4 text-success"
        />
      </template>
      <template v-else-if="resultStatus === PlanCheckRun_Result_Status.WARNING">
        <heroicons-outline:exclamation
          class="-ml-1 mr-1.5 mt-0.5 h-4 w-4 text-warning"
        />
      </template>
      <template v-else-if="resultStatus === PlanCheckRun_Result_Status.ERROR">
        <span class="mr-1.5 font-medium text-error" aria-hidden="true">
          !
        </span>
      </template>
    </template>
    <template v-else-if="status === PlanCheckRun_Status.FAILED">
      <span class="mr-1.5 font-medium text-error" aria-hidden="true"> ! </span>
    </template>
    <template v-else-if="status === PlanCheckRun_Status.CANCELED">
      <heroicons-outline:ban class="-ml-1 mr-1.5 mt-0.5 h-4 w-4 text-control" />
    </template>

    <span>{{ name }}</span>
  </button>
</template>

<script setup lang="ts">
import { maxBy } from "lodash-es";
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { TaskSpinner } from "@/components/IssueV1/components/common";
import type { PlanCheckRun } from "@/types/proto/v1/plan_service";
import {
  PlanCheckRun_Result_Status,
  PlanCheckRun_Status,
  PlanCheckRun_Type,
} from "@/types/proto/v1/plan_service";
import { extractPlanCheckRunUID } from "@/utils";
import { planCheckRunResultStatus } from "./common";

const props = defineProps<{
  planCheckRunList: PlanCheckRun[];
  type: PlanCheckRun_Type;
  clickable?: boolean;
  selected?: boolean;
}>();

defineEmits<{
  (event: "click"): void;
}>();

const { t } = useI18n();

const latestPlanCheckRun = computed(() => {
  // Get the latest PlanCheckRun by UID.
  return maxBy(props.planCheckRunList, (check) =>
    Number(extractPlanCheckRunUID(check.name))
  )!;
});

const status = computed(() => {
  return latestPlanCheckRun.value.status;
});

const resultStatus = computed(() => {
  return planCheckRunResultStatus(latestPlanCheckRun.value);
});

const buttonClasses = computed(() => {
  let bgColor = "";
  let textColor = "";
  let borderColor = "";
  switch (status.value) {
    case PlanCheckRun_Status.RUNNING:
      bgColor = "bg-blue-100";
      textColor = "text-blue-800";
      borderColor = "border-blue-800";
      break;
    case PlanCheckRun_Status.FAILED:
      bgColor = "bg-red-100";
      textColor = "text-red-800";
      borderColor = "border-red-800";
      break;
    case PlanCheckRun_Status.CANCELED:
      bgColor = "bg-yellow-100";
      textColor = "text-yellow-800";
      borderColor = "border-yellow-800";
      break;
    case PlanCheckRun_Status.DONE:
      switch (resultStatus.value) {
        case PlanCheckRun_Result_Status.SUCCESS:
          bgColor = "bg-gray-100";
          textColor = "text-gray-800";
          borderColor = "border-gray-800";
          break;
        case PlanCheckRun_Result_Status.WARNING:
          bgColor = "bg-yellow-100";
          textColor = "text-yellow-800";
          borderColor = "border-yellow-800";
          break;
        case PlanCheckRun_Result_Status.ERROR:
          bgColor = "bg-red-100";
          textColor = "text-red-800";
          borderColor = "border-red-800";
          break;
      }
      break;
  }

  const styleList: string[] = [textColor];
  if (props.clickable) {
    styleList.push("cursor-pointer");
    if (props.selected) {
      styleList.push("font-medium", borderColor);
    } else {
      styleList.push(bgColor, "hover:opacity-80", "border-transparent");
    }
  } else {
    styleList.push(bgColor);
    styleList.push("cursor-default");
  }
  styleList.push("cursor-pointer");
  styleList.push(bgColor);

  return styleList.join(" ");
});

const name = computed(() => {
  const { type } = latestPlanCheckRun.value;
  switch (type) {
    case PlanCheckRun_Type.DATABASE_STATEMENT_FAKE_ADVISE:
      return t('task.check-type.fake');
    case PlanCheckRun_Type.DATABASE_STATEMENT_ADVISE:
      return t('task.check-type.sql-review');
    case PlanCheckRun_Type.DATABASE_CONNECT:
      return t('task.check-type.connection');
    case PlanCheckRun_Type.DATABASE_GHOST_SYNC:
      return t('task.check-type.ghost-sync');
    case PlanCheckRun_Type.DATABASE_STATEMENT_SUMMARY_REPORT:
      return t('task.check-type.summary-report');
    default:
      console.assert(false, `Missing PlanCheckType name of "${type}"`);
      return type;
  }
});
</script>
