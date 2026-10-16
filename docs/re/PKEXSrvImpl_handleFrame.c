
code * FUN_0808cf6a(undefined1 *param_1,uint param_2,int param_3)

{
  char cVar1;
  undefined1 uVar2;
  ostream *poVar3;
  Timer *pTVar4;
  undefined4 uVar5;
  Timer *extraout_ECX;
  Timer *extraout_ECX_00;
  Timer *extraout_ECX_01;
  Timer *extraout_ECX_02;
  Timer *extraout_ECX_03;
  int iVar6;
  Timer *extraout_ECX_04;
  Timer **ppTVar7;
  undefined1 *puVar8;
  Timer **ppTVar9;
  uint *puVar10;
  code *pcVar11;
  byte bVar12;
  char *pcVar13;
  Timer *pTVar14;
  Timer *pTStack_90;
  undefined1 auStack_8c [12];
  Timer *local_80;
  Timer *local_7c;
  undefined4 local_78;
  undefined4 *local_74;
  int local_70;
  undefined4 local_6c;
  undefined4 local_68;
  undefined4 local_60;
  undefined4 *local_5c;
  code *local_58;
  undefined4 local_54;
  undefined4 local_50;
  undefined4 *local_4c;
  code *local_48;
  undefined4 local_44;
  undefined4 local_40;
  undefined4 *local_3c;
  code *local_38;
  undefined4 local_34;
  
  bVar12 = 0;
  pTVar4 = *(Timer **)(param_1 + 200);
  if (pTVar4 == (Timer *)0x5) {
    return (code *)0x0;
  }
  if (param_2 == 0x45) {
    local_58 = *(code **)(param_3 + 8);
    local_60 = 0;
    local_5c = *(undefined4 **)(param_3 + 4);
    local_54 = *(undefined4 *)(param_3 + 0xc);
    if (pTVar4 == (Timer *)0x4) {
      local_40 = 0;
      pTStack_90 = (Timer *)0x0;
      nv::Timer::operator=((Timer *)(param_1 + 0x50),(Timer *)&local_40);
      nv::Timer::~Timer((Timer *)&local_40);
      *(int *)(param_1 + 0x54) = 0;
      local_6c = 6;
      FUN_0805e79c();
      local_40 = 0;
      local_38 = local_58;
      local_3c = local_5c;
      local_34 = local_54;
      uVar2 = FUN_0808669c();
      local_80 = (Timer *)CONCAT31(local_80._1_3_,uVar2);
      FUN_08050ec6();
      FUN_0805e336();
      ppTVar7 = (Timer **)auStack_8c;
      if ((char)local_80 == '\0') {
        local_74 = (undefined4 *)0x0;
        local_70 = 0;
        pTStack_90 = (Timer *)0x808d53e;
        FUN_08051b82();
        cVar1 = FUN_08086816();
        if (cVar1 == '\0') {
          pTStack_90 = *(Timer **)param_1;
          pcVar11 = *(code **)(pTStack_90 + 0x10);
          string::string((string *)&local_40,"auth failed (4)");
          (*pcVar11)();
          string::freeptr();
          ppTVar9 = (Timer **)auStack_8c;
        }
        else {
          local_50 = 0;
          local_48 = (code *)(local_70 - (int)local_74);
          local_4c = local_74;
          local_44 = 0;
          pTStack_90 = (Timer *)0x808d5aa;
          FUN_0805e77a();
          local_40 = 0;
          local_38 = local_48;
          local_3c = local_4c;
          local_34 = local_44;
          pTStack_90 = (Timer *)0x808d5ca;
          uVar2 = FUN_0808669c();
          local_80 = (Timer *)CONCAT31(local_80._1_3_,uVar2);
          pTStack_90 = (Timer *)0x808d5d4;
          FUN_08050ec6();
          FUN_0805e336();
          if ((char)local_80 == '\0') {
            pTStack_90 = (Timer *)0x808d623;
            FUN_080882fc();
            pTStack_90 = (Timer *)0x808d631;
            FUN_08086d64();
            local_78 = 0;
            pTStack_90 = (Timer *)0x808d63e;
            FUN_08086d8a();
            local_80 = *(Timer **)(param_1 + 0xac);
            pTStack_90 = (Timer *)0x808d656;
            FUN_08051b82();
            cVar1 = FUN_08086eda();
            if (cVar1 == '\0') {
LAB_0808d670:
              pTStack_90 = *(Timer **)param_1;
              pcVar11 = *(code **)(pTStack_90 + 0x10);
              goto LAB_0808d677;
            }
            pTStack_90 = (Timer *)0x808d690;
            FUN_08051b82();
            pTStack_90 = (Timer *)0x808d69d;
            cVar1 = FUN_08089ce6();
            if (cVar1 == '\0') goto LAB_0808d670;
            pTStack_90 = (Timer *)0x808d6b0;
            FUN_08051b82();
            poVar3 = (ostream *)operator<<();
            local_80 = (Timer *)operator<<(poVar3,": peer id: ");
            FUN_08086420();
            poVar3 = (ostream *)FUN_08051836();
            endl(poVar3);
            string::freeptr();
            local_80 = *(Timer **)(param_1 + 0xc4);
            if (local_80 == (Timer *)0x0) {
              pcVar11 = *(code **)(*(int *)param_1 + 0x10);
              pcVar13 = "no peer";
              pTStack_90 = extraout_ECX_03;
              goto LAB_0808d71a;
            }
            pTStack_90 = (Timer *)0x808d737;
            FUN_08087ee2();
            pTVar4 = (Timer *)&local_40;
            pTVar14 = local_80;
            for (iVar6 = 8; iVar6 != 0; iVar6 = iVar6 + -1) {
              *(undefined4 *)pTVar14 = *(undefined4 *)pTVar4;
              pTVar4 = pTVar4 + (uint)bVar12 * -8 + 4;
              pTVar14 = pTVar14 + (uint)bVar12 * -8 + 4;
            }
            pTStack_90 = (Timer *)0x808d74e;
            iVar6 = FUN_08051968();
            if (iVar6 == 2) {
              pTVar14 = (Timer *)0x8095141;
              pTStack_90 = extraout_ECX_04;
              poVar3 = (ostream *)operator<<();
              poVar3 = (ostream *)operator<<(poVar3,": received addrs");
              endl(poVar3);
              m3::message::read::get((uint)&local_40);
              local_7c = (Timer *)param_1;
              m3::each(local_40);
              ppTVar7 = &pTStack_90;
            }
            pcVar11 = *(code **)(*(int *)param_1 + 0x10);
            *(Timer **)((int)ppTVar7 + -4) = pTVar14;
            *(Timer **)((int)ppTVar7 + -8) = pTVar14;
            *(undefined **)((int)ppTVar7 + -0xc) = &DAT_08092eed;
            *(undefined4 **)((int)ppTVar7 + -0x10) = &local_40;
            *(undefined4 *)((int)ppTVar7 + -0x14) = 0x808d7c4;
            string::string(*(string **)((int)ppTVar7 + -0x10),*(char **)((int)ppTVar7 + -0xc));
            *(undefined4 **)((int)ppTVar7 + -0xc) = &local_40;
            puVar8 = (undefined1 *)((int)ppTVar7 + -0x10);
            *(undefined1 **)((int)ppTVar7 + -0x10) = param_1;
            *(undefined4 *)((int)ppTVar7 + -0x14) = 0x808d7ca;
            (*pcVar11)();
            *(undefined4 **)((int)ppTVar7 + -0x10) = &local_40;
          }
          else {
            pTStack_90 = extraout_ECX_02;
            poVar3 = (ostream *)operator<<();
            poVar3 = (ostream *)operator<<(poVar3,": missing fields in decrypted msg");
            endl(poVar3);
            pcVar11 = *(code **)(*(int *)param_1 + 0x10);
LAB_0808d677:
            pcVar13 = "auth failed (4)";
LAB_0808d71a:
            string::string((string *)&local_40,pcVar13);
            puVar8 = &stack0xffffff64;
            (*pcVar11)();
          }
          *(undefined4 *)(puVar8 + -4) = 0x808d7d2;
          string::freeptr();
          ppTVar9 = (Timer **)(puVar8 + 0x10);
          *(undefined4 *)(puVar8 + 0xc) = 0x808d7dd;
          FUN_08050ec6();
        }
        *(undefined4 *)((int)ppTVar9 + -4) = 0x808d7e5;
        FUN_0805da10();
LAB_0808d7e5:
        pTVar4 = (Timer *)(code *)0x1;
        goto LAB_0808d7ed;
      }
      pTStack_90 = (Timer *)param_1;
      poVar3 = (ostream *)operator<<();
      pcVar13 = ": invalid msg4 received";
LAB_0808d50f:
      poVar3 = (ostream *)operator<<(poVar3,pcVar13);
      endl(poVar3);
    }
LAB_0808d520:
    ppTVar9 = (Timer **)auStack_8c;
    pTVar4 = (Timer *)(code *)0x0;
  }
  else {
    if (0x45 < param_2) {
LAB_0808d7fa:
      pTStack_90 = pTVar4;
      uVar5 = operator<<();
      puVar10 = (uint *)&stack0xffffff68;
      goto LAB_0808d838;
    }
    if (param_2 != 0x40) {
      if (param_2 != 0x43) goto LAB_0808d7fa;
      local_58 = *(code **)(param_3 + 8);
      local_60 = 0;
      local_5c = *(undefined4 **)(param_3 + 4);
      local_54 = *(undefined4 *)(param_3 + 0xc);
      if (pTVar4 != (Timer *)0x3) goto LAB_0808d520;
      local_40 = 0;
      pTStack_90 = (Timer *)0x0;
      nv::Timer::operator=((Timer *)(param_1 + 0x50),(Timer *)&local_40);
      nv::Timer::~Timer((Timer *)&local_40);
      *(int *)(param_1 + 0x54) = 0;
      local_6c = 0;
      local_68 = 4;
      FUN_0805e79c();
      local_40 = 0;
      local_38 = local_58;
      local_3c = local_5c;
      local_34 = local_54;
      cVar1 = FUN_0808669c();
      FUN_08050ec6();
      FUN_0805e336();
      if (cVar1 != '\0') {
        pTStack_90 = (Timer *)param_1;
        poVar3 = (ostream *)operator<<();
        pcVar13 = "invalid msg2 received";
        goto LAB_0808d50f;
      }
      pTStack_90 = (Timer *)0x808d345;
      FUN_080882fc();
      pTStack_90 = (Timer *)0x808d353;
      FUN_08086d64();
      local_6c = 0;
      pTStack_90 = (Timer *)0x808d360;
      FUN_08086d8a();
      pTStack_90 = (Timer *)0x808d375;
      FUN_08051b82();
      pTVar4 = (Timer *)FUN_08086eda();
      if ((char)pTVar4 == '\0') {
        pTStack_90 = *(Timer **)param_1;
        pcVar11 = *(code **)(pTStack_90 + 0x10);
        string::string((string *)&local_40,"auth failed (2)");
        (*pcVar11)();
        string::freeptr();
        ppTVar9 = (Timer **)auStack_8c;
        goto LAB_0808d7e5;
      }
      local_80 = *(Timer **)(param_1 + 0x58);
      pTStack_90 = (Timer *)0x808d3c5;
      pTStack_90 = (Timer *)FUN_0808a80d();
      WCurve::fromEC((ECPoint *)&local_40);
      ppTVar9 = &pTStack_90;
      uVar2 = FUN_0808bfd4();
      local_80 = (Timer *)CONCAT31(local_80._1_3_,uVar2);
      FUN_08086dc4();
      FUN_08086d42();
      if ((char)local_80 == '\0') {
        local_80 = *(Timer **)(*(int *)param_1 + 0x10);
        string::string((string *)&local_40,"auth failed (2)");
        (*(code *)local_80)(param_1);
        string::freeptr();
        ppTVar9 = &pTStack_90;
      }
      else {
        FUN_08089794();
        FUN_08086d64();
        FUN_0808cce6();
        pTVar4 = local_80;
      }
      goto LAB_0808d7ed;
    }
    pTStack_90 = (Timer *)0x808cfbc;
    FUN_08051b82();
    if ((char)local_58 == '\0') {
      return local_58;
    }
    pTStack_90 = (Timer *)0x808cfd6;
    pTVar4 = (Timer *)FUN_0808c8d2();
    if ((char)pTVar4 == '\0') {
      *(int *)(param_1 + 200) = 5;
      pTStack_90 = *(Timer **)(*(int *)param_1 + 0x10);
      local_80 = pTStack_90;
      string::string((string *)&local_40,"session overlap detected");
      (*(code *)local_80)();
      string::freeptr();
      return (code *)pTVar4;
    }
    local_48 = *(code **)(param_3 + 8);
    local_50 = 0;
    local_4c = *(undefined4 **)(param_3 + 4);
    local_44 = *(undefined4 *)(param_3 + 0xc);
    if (*(uint *)(param_1 + 200) < 3) {
      local_74 = (undefined4 *)0x1;
      pTStack_90 = (Timer *)0x808d054;
      FUN_0805e79c();
      local_40 = 0;
      local_38 = local_48;
      local_3c = local_4c;
      local_34 = local_44;
      pTStack_90 = (Timer *)0x808d075;
      uVar2 = FUN_0808669c();
      local_80 = (Timer *)CONCAT31(local_80._1_3_,uVar2);
      pTStack_90 = (Timer *)0x808d080;
      FUN_08050ec6();
      FUN_0805e336();
      if ((char)local_80 != '\0') {
        param_1 = cout;
        pTStack_90 = extraout_ECX;
        poVar3 = (ostream *)operator<<();
        poVar3 = (ostream *)operator<<(poVar3,": invalid discovery request frame received");
        endl(poVar3);
        goto LAB_0808d0c1;
      }
      if (*(int *)(param_1 + 200) == 0) {
        poVar3 = (ostream *)operator<<();
        poVar3 = (ostream *)operator<<(poVar3,": session started");
        endl(poVar3);
        local_80 = *(Timer **)(param_1 + 0x10);
        *(int *)(param_1 + 200) = 1;
        local_40 = 0;
        local_3c = malloc(8);
        *local_3c = &PTR_LAB_080989e4;
        local_3c[1] = param_1;
        nv::Timer::Timer();
        nv::Timer::operator=((Timer *)(param_1 + 0xd0),(Timer *)&local_6c);
        nv::Timer::~Timer((Timer *)&local_6c);
        FUN_08051296();
      }
      pTStack_90 = (Timer *)&local_6c;
      poVar3 = (ostream *)operator<<();
      endl(poVar3);
      pTStack_90 = extraout_ECX_00;
      if (DAT_0809a8a8 == 1) {
        if (((byte)runtimeDFLAGS & 2) != 0) goto LAB_0808d199;
      }
      else if (-1 < DAT_0809a8a8 + -1) {
LAB_0808d199:
        if (2 < *(uint *)(param_1 + 200)) {
          pTStack_90 = (Timer *)&local_6c;
          poVar3 = (ostream *)operator<<((ostream *)cerr,"\x1b[1m");
          poVar3 = (ostream *)
                   operator<<(poVar3,
                              "\"pkex/pkex.cpp\":445: ECF\x1b[22m oneof(state, Discovery1, Discovery2, Discovery3): "
                             );
          endl(poVar3);
          FUN_080865c4();
          pTStack_90 = extraout_ECX_01;
        }
      }
      if (*(int *)(param_1 + 200) == 2) {
        poVar3 = (ostream *)operator<<();
        poVar3 = (ostream *)operator<<(poVar3,": discovery finished");
        endl(poVar3);
        FUN_0808ca5a();
      }
      else {
        local_40 = 0;
        pTStack_90 = (Timer *)0x808d1f8;
        FUN_0805d9b6();
        pTStack_90 = (Timer *)0x808d20e;
        FUN_0805d9b6();
        pTStack_90 = (Timer *)0x808d221;
        FUN_0805d9b6();
        FUN_08089616();
        FUN_08051114();
      }
    }
    else {
LAB_0808d0c1:
      pTVar4 = (Timer *)0x0;
    }
    ppTVar9 = (Timer **)auStack_8c;
  }
LAB_0808d7ed:
  *(undefined4 *)((int)ppTVar9 + -4) = 0x808d7f2;
  FUN_08050ec6();
  if ((char)pTVar4 != '\0') {
    return (code *)pTVar4;
  }
  *(undefined1 **)((int)ppTVar9 + -4) = param_1;
  *(undefined4 *)((int)ppTVar9 + -8) = 0x1d;
  *(char **)((int)ppTVar9 + -0xc) =
       "bool PKEXSrvImpl::handleFrame(unsigned int, m3::msg_view, const PKEX::Context&)";
  *(undefined1 **)((int)ppTVar9 + -0x10) = cout;
  *(undefined4 *)((int)ppTVar9 + -0x14) = 0x808d831;
  uVar5 = operator<<();
  puVar10 = (uint *)((int)ppTVar9 + -0xc);
  *(char **)((int)ppTVar9 + -0xc) = ": unhandled msg ";
LAB_0808d838:
  puVar10[-1] = uVar5;
  puVar10[-2] = 0x808d83e;
  uVar5 = operator<<((ostream *)puVar10[-1],(char *)*puVar10);
  *puVar10 = param_2;
  puVar10[-1] = uVar5;
  puVar10[-2] = 0x808d849;
  uVar5 = operator<<((ostream *)puVar10[-1],*puVar10);
  puVar10[-1] = uVar5;
  puVar10[-2] = 0x808d851;
  endl((ostream *)puVar10[-1]);
  return (code *)0x0;
}


