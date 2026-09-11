import { apiClient } from './client';

export type TaskRelationType = 'SS' | 'FS' | 'FF' | 'SF';

export interface TaskDependency {
  id: number;
  task_id: number;
  depends_on_task_id: number;
  relation_type: TaskRelationType;
  lag_days: number;
}

export async function listTaskDependencies(projectId: number): Promise<TaskDependency[]> {
  const { data } = await apiClient.get<TaskDependency[]>(`/api/v1/projects/${projectId}/dependencies`);
  return data;
}

export async function createTaskDependency(
  projectId: number,
  taskId: number,
  dependsOnTaskId: number,
  relationType: TaskRelationType,
  lagDays: number,
): Promise<TaskDependency> {
  const { data } = await apiClient.post<TaskDependency>(`/api/v1/projects/${projectId}/dependencies`, {
    task_id: taskId,
    depends_on_task_id: dependsOnTaskId,
    relation_type: relationType,
    lag_days: lagDays,
  });
  return data;
}

export async function updateTaskDependency(
  projectId: number,
  dependencyId: number,
  relationType: TaskRelationType,
  lagDays: number,
): Promise<TaskDependency> {
  const { data } = await apiClient.put<TaskDependency>(
    `/api/v1/projects/${projectId}/dependencies/${dependencyId}`,
    { relation_type: relationType, lag_days: lagDays },
  );
  return data;
}

export async function deleteTaskDependency(projectId: number, dependencyId: number): Promise<void> {
  await apiClient.delete(`/api/v1/projects/${projectId}/dependencies/${dependencyId}`);
}
