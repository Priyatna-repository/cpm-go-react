import { apiClient } from './client';
import type { ProjectUserRef, ProjectLabelRef } from './projects';

export type TaskType = 'task' | 'summary' | 'milestone';

export interface Task {
  id: number;
  project_id: number;
  type: TaskType;
  parent_task_id?: number;
  name: string;
  description?: string;
  number: number;
  order_column: number;
  assigned_to_user?: ProjectUserRef;
  start_date?: string;
  end_date?: string;
  budget_plan?: number;
  budget_actual?: number;
  weight: number;
  progress: number;
  status_label?: ProjectLabelRef;
  created_by?: ProjectUserRef;
  created_at: string;
  is_completed: boolean;
  completed_at?: string;
  archived_at?: string;
}

export interface TaskInput {
  type: TaskType;
  parent_task_id?: number | null;
  name: string;
  description?: string | null;
  assigned_to_user_id?: number | null;
  start_date?: string | null;
  end_date?: string | null;
  budget_plan?: number | null;
  budget_actual?: number | null;
  weight: number;
  status_label_id?: number | null;
}

export async function listTasks(projectId: number, parentTaskId?: number, archived = false): Promise<Task[]> {
  const { data } = await apiClient.get<Task[]>(`/api/v1/projects/${projectId}/tasks`, {
    params: { parent_task_id: parentTaskId, archived: archived ? 1 : undefined },
  });
  return data;
}

export async function createTask(projectId: number, input: TaskInput): Promise<Task> {
  const { data } = await apiClient.post<Task>(`/api/v1/projects/${projectId}/tasks`, input);
  return data;
}

export async function updateTask(projectId: number, id: number, input: TaskInput): Promise<Task> {
  const { data } = await apiClient.put<Task>(`/api/v1/projects/${projectId}/tasks/${id}`, input);
  return data;
}

export async function archiveTask(projectId: number, id: number): Promise<void> {
  await apiClient.delete(`/api/v1/projects/${projectId}/tasks/${id}`);
}

export async function restoreTask(projectId: number, id: number): Promise<void> {
  await apiClient.post(`/api/v1/projects/${projectId}/tasks/${id}/restore`);
}

export async function forceDeleteTask(projectId: number, id: number): Promise<void> {
  await apiClient.delete(`/api/v1/projects/${projectId}/tasks/${id}/force`);
}

// reschedule triggers the server-side cascade (TaskSchedulingService) and
// returns EVERY task whose dates changed as a result — not just the one
// that was dragged. The caller must push every entry in the response back
// into whatever's rendering the Gantt, not just the dragged task.
export async function rescheduleTask(
  projectId: number,
  id: number,
  startDate: string | null,
  endDate: string | null,
): Promise<Task[]> {
  const { data } = await apiClient.put<Task[]>(`/api/v1/projects/${projectId}/tasks/${id}/reschedule`, {
    start_date: startDate,
    end_date: endDate,
  });
  return data;
}

export async function setTaskCompleted(projectId: number, id: number, completed: boolean): Promise<Task> {
  const { data } = await apiClient.put<Task>(`/api/v1/projects/${projectId}/tasks/${id}/complete`, { completed });
  return data;
}
