import { zodResolver } from "@hookform/resolvers/zod"
import { useEffect } from "react"
import { useForm } from "react-hook-form"
import { z } from "zod"

import { Button } from "@workspace/ui/components/button"
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@workspace/ui/components/dialog"
import { Label } from "@workspace/ui/components/label"
import { Select } from "@workspace/ui/components/select"
import { Textarea } from "@workspace/ui/components/textarea"

import { cardLimitLabel, cardTypeLabel } from "@/lib/format"
import type { CardImportPayload } from "@/lib/types"

const importSchema = z.object({
  rawText: z.string().trim().min(1, "请至少输入一个卡密"),
  cardType: z.enum(["us", "uk", "es"], {
    message: "请选择卡片类型",
  }),
  cardLimit: z.enum(["0", "1", "2"], {
    message: "请选择额度",
  }),
})

type ImportValues = z.infer<typeof importSchema>

const DEFAULT_IMPORT_VALUES: ImportValues = {
  rawText: "",
  cardType: "uk",
  cardLimit: "0",
}

export function CardImportDialog({
  open,
  onOpenChange,
  onSubmit,
  isPending,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (payload: CardImportPayload) => Promise<unknown>
  isPending: boolean
}) {
  const form = useForm<ImportValues>({
    resolver: zodResolver(importSchema),
    defaultValues: DEFAULT_IMPORT_VALUES,
  })

  useEffect(() => {
    if (!open) {
      form.reset(DEFAULT_IMPORT_VALUES)
    }
  }, [form, open])

  const handleSubmit = form.handleSubmit(async (values) => {
    await onSubmit({
      rawText: values.rawText,
      cardType: values.cardType,
      cardLimit: Number(values.cardLimit) as CardImportPayload["cardLimit"],
    })
    form.reset(DEFAULT_IMPORT_VALUES)
  })

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>批量导入卡密</DialogTitle>
          <DialogDescription>
            一行一个卡密。去重、空白过滤和已存在卡密过滤都在后端完成。
          </DialogDescription>
        </DialogHeader>
        <DialogBody className="space-y-4">
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="card-import-type">卡片类型</Label>
              <Select id="card-import-type" {...form.register("cardType")}>
                <option value="">请选择</option>
                <option value="us">{cardTypeLabel("us")}</option>
                <option value="uk">{cardTypeLabel("uk")}</option>
                <option value="es">{cardTypeLabel("es")}</option>
              </Select>
              {form.formState.errors.cardType ? (
                <p className="text-sm text-rose-600">
                  {form.formState.errors.cardType.message}
                </p>
              ) : null}
            </div>

            <div className="space-y-2">
              <Label htmlFor="card-import-limit">额度</Label>
              <Select id="card-import-limit" {...form.register("cardLimit")}>
                <option value="">请选择</option>
                <option value="0">{cardLimitLabel(0)}</option>
                <option value="1">{cardLimitLabel(1)}</option>
                <option value="2">{cardLimitLabel(2)}</option>
              </Select>
              {form.formState.errors.cardLimit ? (
                <p className="text-sm text-rose-600">
                  {form.formState.errors.cardLimit.message}
                </p>
              ) : null}
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="card-import-textarea">卡密内容</Label>
            <Textarea
              id="card-import-textarea"
              placeholder={"CDK-ABC12345\nCDK-DEF67890"}
              {...form.register("rawText")}
            />
            {form.formState.errors.rawText ? (
              <p className="text-sm text-rose-600">
                {form.formState.errors.rawText.message}
              </p>
            ) : null}
          </div>
        </DialogBody>
        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
          >
            取消
          </Button>
          <Button type="button" onClick={() => void handleSubmit()} disabled={isPending}>
            {isPending ? "导入中..." : "开始导入"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
