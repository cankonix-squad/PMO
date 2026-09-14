"use client";
import type { QueryClient } from "@tanstack/react-query";
import type { Dispatch, SetStateAction } from "react";
import { projectService } from "@/services/project.service";
import type { CreateVendorRequest, Vendor } from "@/types/project";
import { useMutation } from "@tanstack/react-query";

interface Options {
  setVendorFormOpen: Dispatch<SetStateAction<boolean>>;
  setEditingVendor: Dispatch<SetStateAction<Vendor | null>>;
  setFormError: Dispatch<SetStateAction<string | null>>;
  qc: QueryClient;
}

export function useVendorMutations({ setVendorFormOpen, setEditingVendor, setFormError, qc }: Options) {
  const createVendorMutation = useMutation({
    mutationFn: (payload: CreateVendorRequest) =>
      projectService.createVendor(payload),
    onSuccess: () => {
      setVendorFormOpen(false);
      setEditingVendor(null);
      setFormError(null);
      void qc.invalidateQueries({ queryKey: ["vendors", "all"] });
    },
  });
  return { createVendorMutation };
}
