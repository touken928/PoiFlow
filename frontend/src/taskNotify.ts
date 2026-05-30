import {
    CheckNotificationAuthorization,
    InitializeNotifications,
    IsNotificationAvailable,
    RequestNotificationAuthorization,
    SendNotification,
} from '../wailsjs/runtime/runtime';

let ready = false;

async function ensureNotifications(): Promise<boolean> {
    if (ready) return true;
    try {
        if (!(await IsNotificationAvailable())) return false;
        await InitializeNotifications();
        if (!(await CheckNotificationAuthorization())) {
            await RequestNotificationAuthorization();
        }
        ready = true;
        return true;
    } catch {
        return false;
    }
}

interface TaskBrief {
    id: string;
    name: string;
    records?: number;
}

export async function notifyTaskCompleted(task: TaskBrief, recordCount: number) {
    if (!(await ensureNotifications())) return;
    try {
        await SendNotification({
            id: `task-done-${task.id}`,
            title: '采集完成',
            subtitle: task.name,
            body: `共 ${recordCount} 条 POI 数据`,
        });
    } catch {
        /* ignore */
    }
}

export async function notifyTaskFailed(task: TaskBrief, error?: string) {
    if (!(await ensureNotifications())) return;
    try {
        await SendNotification({
            id: `task-fail-${task.id}`,
            title: '采集失败',
            subtitle: task.name,
            body: error || '任务执行出错',
        });
    } catch {
        /* ignore */
    }
}

export async function notifyTaskPaused(task: TaskBrief, reason: string) {
    if (!(await ensureNotifications())) return;
    try {
        await SendNotification({
            id: `task-pause-${task.id}`,
            title: '任务已暂停',
            subtitle: task.name,
            body: reason,
        });
    } catch {
        /* ignore */
    }
}
